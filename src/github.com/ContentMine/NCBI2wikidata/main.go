//   Copyright 2019 Content Mine Ltd
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ContentMine/EUtils"
	europmc "github.com/ContentMine/go-europmc"
	"github.com/jlaffaye/ftp"
)

const EFETCH_BATCH_SIZE = 200

const NCBI_LICENSE_URL = "ftp://ftp.ncbi.nlm.nih.gov:21/pub/pmc/oa_file_list.txt"
const NCBI_FILE_FILE = "oa_file_list.txt"

func FetchLicenses(target_filename string, ftp_location string) error {
	url, err := url.Parse(ftp_location)
	if err != nil {
		return err
	}

	if url.Scheme != "ftp" {
		return fmt.Errorf("We require an FTP URL, not %s", ftp_location)
	}

	client, err := ftp.Dial(url.Host)
	if err != nil {
		return err
	}

	err = client.Login("anonymous", "anonymous")
	if err != nil {
		return err
	}

	resp, err := client.Retr(url.Path)
	if err != nil {
		return err
	}
	defer resp.Close()

	f, err := os.Create(target_filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp)
	return err
}

func set_to_list(m map[string]string) []string {
	r := make([]string, len(m))
	i := 0
	for k, _ := range m {
		r[i] = k
		i += 1
	}
	return r
}

// Load the NCBI open access file list so we can map PMID -> Copyright
//
// First line is date file was generated, rest are tab separated info on papers. Example:
// oa_package/87/30/PMC17774.tar.gz	Arthritis Res. 1999 Oct 14; 1(1):63-70	PMC17774	PMID:11056661	NO-CC CODE
//
func LoadLicenses(filename string, license_map map[string]string) error {

	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// File just not there, so try to fetch it first
			log.Printf("Fetching PMC open access list, this may take some time...")
			err := FetchLicenses(filename, NCBI_LICENSE_URL)
			if err != nil {
				return err
			}
			log.Printf("Fetching PMC open access list complete.")
			// Now try again
			f, err = os.Open(filename)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	var line string
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}

		parts := strings.Split(line, "\t")
		if len(parts) != 5 {
			continue
		}

		var pmid string
		prefixed_pmid := parts[3]
		split_pmid := strings.Split(prefixed_pmid, ":")
		if len(split_pmid) == 2 {
			pmid = split_pmid[1]
		}

		pmcid := parts[3]

		license := strings.TrimSuffix(parts[4], "\n")

		// if PMID is in target list store info
		if _, ok := license_map[pmid]; ok {
			license_map[pmid] = license
		}

		if _, ok := license_map[pmcid]; ok {
			license_map[pmcid] = license
		}
	}

	return nil
}

type Record struct {
	Title           string
	MainSubjects    []EUtils.MeshDescriptorName
	IsReview        bool
	PublicationDate string
	Publication     string
	ISSN            string
	PMCLicense      string
	EPMCLicenseLink string
	PMID            string
	PMCID           string
	IsRetracted     bool
	IsRetraction    bool
	RetractedByPMID string
}

func GetEuroPMCLicenseLinkForPMCID(pmcid string) (string, error) {

	// Go ask EPMC about the license to get more details
	paper, err := europmc.FetchFullText(pmcid)
	if err != nil {
		return "", err
	}
	license_info := paper.Front.ArticleMeta.Permissions.License
	if license_info.Link == "" {
		if strings.Contains(license_info.Text, "This article is distributed under the terms of the Creative Commons Attribution 4.0 International License") {
			license_info.Link = "https://creativecommons.org/licenses/by/4.0/"
		}
	} else {
		// The URLs between wikidata and EPMC aren't very consistent: some are HTTP, some HTTPS, some
		// have a training / some do not, etc. So we try to move to a canonical form here
		u, err := url.Parse(license_info.Link)
		if err != nil {
			log.Printf("Failed to parse license link %s: %s", license_info.Link, err)
		} else {
			u.Scheme = "https"
			if !strings.HasSuffix(u.Path, "/") {
				u.Path += "/"
			}
			license_info.Link = u.String()
		}
	}

	return license_info.Link, nil
}

func ArticleToRecord(article EUtils.PubmedArticle) Record {

	return Record{
		Title:           article.MedlineCitation.Article[0].ArticleTitle,
		PMID:            article.MedlineCitation.PMID,
		PMCID:           article.GetPMCID(),
		PMCLicense:      "",
		MainSubjects:    article.GetMajorTopics(),
		PublicationDate: article.GetPublicationDateString(),
		Publication:     article.MedlineCitation.Article[0].Journal.Title,
		ISSN:            article.MedlineCitation.Article[0].Journal.ISSN,
		IsReview:        article.IsReview(),
		IsRetracted:     article.IsRetracted(),
		IsRetraction:    article.IsRetraction(),
		RetractedByPMID: article.GetRetractedInPMID(),
	}
}

func batch(term string, ncbi_api_key string, csv_file *os.File, qs_file *os.File) error {

	// Because we use the history feature of the eUtilities API, it doesn't matter how many
	// things get returned here, we rely on the eFetch API to get all the deets. Hence the
	// single request here. We really are just doing this to light up things later
	search_request := EUtils.ESearchRequest{
		DB:         "pubmed",
		APIKey:     os.Getenv("NCBI_API_KEY"),
		Term:       term,
		RetMax:     1,
		UseHistory: true,
	}

	search_resp, err := search_request.Do()
	if err != nil {
		return err
	}

	count, err := strconv.Atoi(search_resp.Count)
	if err != nil {
		return err
	}
	log.Printf("Search returned %d matches for %s.\n", count, term)

	// Things to build up as we fetch the results from PMC...
	all_records := make([]Record, 0)
	pmid_set := make(map[string]string, 0)
	pmcid_set := make(map[string]string, 0)
	issn_set := make(map[string]string, 0)
	main_subject_set := make(map[string]string, 0)
	license_map := make(map[string]string, 0)

	for i := 0; i < count; i += EFETCH_BATCH_SIZE {

		fetch_request := EUtils.EFetchHistoryRequest{
			DB:       "pubmed",
			WebEnv:   search_resp.WebEnv,
			QueryKey: search_resp.QueryKey,
			APIKey:   ncbi_api_key,
			RetStart: i,
			RetMax:   EFETCH_BATCH_SIZE,
		}
		// this is a lazy way to do rate limiting - we're allowed ten requests on the NCBI API a second. This
		// ensures we'll never hit this. We could do better, but it's not worth the complexity IMHO.
		time.Sleep(100 * time.Millisecond)

		fetch_resp, err := fetch_request.Do()
		if err != nil {
			return err
		}

		log.Printf("Fetched %d articles for %s.\n", len(fetch_resp.Articles), term)

		for _, article := range fetch_resp.Articles {

			// Distill out what we want from the article
			record := ArticleToRecord(article)

			if record.PMID != "" {
				license_map[record.PMID] = ""
			}

			// make a note of the things we need to look up on wikidata
			if record.RetractedByPMID != "" {
				pmid_set[record.RetractedByPMID] = ""
			}
			for _, subject := range record.MainSubjects {
				main_subject_set[subject.MeshID] = ""
			}
			if record.ISSN != "" {
				issn_set[record.ISSN] = ""
			}
			if record.PMCID != "" {
				pmcid_set[record.PMCID] = ""
				license_map[record.PMCID] = ""
			}

			all_records = append(all_records, record)
		}
	}

	err = LoadLicenses(NCBI_FILE_FILE, license_map)

	licensed_records := make([]Record, 0)

	for _, record := range all_records {
		license, ok := license_map[record.PMID]
		if ok && license != "" {
			record.PMCLicense = license
		} else {
			license, ok := license_map[record.PMCID]
			if ok && license != "" {
				record.PMCLicense = license
			} else {
				continue
			}
		}
		licensed_records = append(licensed_records, record)
	}

	log.Printf("We got information on %d records.\n", len(licensed_records))

	pmcid_list := set_to_list(pmcid_set)
	log.Printf("Getting IDs for %d PMCID items", len(pmcid_list))
	pmcid_wikidata_items, err := PMCIDsToWDItem(pmcid_list)
	if err != nil {
		return fmt.Errorf("Failed fetching %d PMCID items: %v", len(pmcid_list), err)
	}
	pmid_list := set_to_list(pmid_set)
	log.Printf("Getting IDs for %d PMID items", len(pmid_list))
	pmid_wikidata_items, err := PMIDsToWDItem(pmid_list)
	if err != nil {
		return fmt.Errorf("Failed fetching %d PMID items: %v", len(pmid_list), err)
	}
	issn_list := set_to_list(issn_set)
	log.Printf("Getting IDs for %d ISSN items", len(issn_list))
	issn_wikidata_items, err := ISSNsToWDItem(issn_list)
	if err != nil {
		return fmt.Errorf("Failed fetching %d ISSN items: %v", len(issn_list), err)
	}
	main_subject_list := set_to_list(main_subject_set)
	log.Printf("Getting IDs for %d drug/disease items", len(main_subject_list))
	drug_wikidata_items, err := DrugsToWDItem(main_subject_list)
	if err != nil {
		return fmt.Errorf("Failed fetching drug %d items: %v", len(main_subject_list), err)
	}
	disease_wikidata_items, err := DiseasesToWDItem(main_subject_list)
	if err != nil {
		return fmt.Errorf("Failed fetching %d disease items: %v", len(main_subject_list), err)
	}

	now := time.Now()

	for _, record := range licensed_records {

		item := pmcid_wikidata_items[record.PMCID]
		issn_item := issn_wikidata_items[record.ISSN]

		// see of we can get better license detail from EuroPMC
		if record.PMCID != "" {
			record.EPMCLicenseLink, err = GetEuroPMCLicenseLinkForPMCID(record.PMCID)
			if err != nil {
				log.Printf("Failed to get EPMC data for %s: %v", record.PMCID, err)
			}
		}

		license_item := CC_LICENSE_ITEM_IDS[record.EPMCLicenseLink]
		license_source := EuroPMC_ITEM
		if license_item == "" {
			license_item = CC_LICENSE_ITEM_IDS[record.PMCLicense]
			license_source = PMC_ITEM
		}

		retracted_by_item := pmid_wikidata_items[record.RetractedByPMID]

		if item != "" {
			statement := AddStringPropertyToItem(item, PMCID_PROPERTY, record.PMCID)
			statement.AddSource(STATED_IN_SOURCE, PMC_ITEM)
			statement.AddSource(RETRIEVED_AT_DATE_SOURCE, fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/11", now.Year(), now.Month(), now.Day()))
			qs_file.WriteString(fmt.Sprintf("%v", statement))

			if record.PublicationDate != "" {
				statement = AddStringPropertyToItem(item, PUBLICATION_DATE_PROPERTY, record.PublicationDate)
				statement.AddSource(STATED_IN_SOURCE, PMC_ITEM)
				statement.AddSource(RETRIEVED_AT_DATE_SOURCE, fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/11", now.Year(), now.Month(), now.Day()))
				qs_file.WriteString(fmt.Sprintf("%v", statement))
			}

			if record.IsReview {
				statement := AddItemPropertyToItem(item, INSTANCE_OF_PROPERTY, REVIEW_ARTICLE_ITEM)
				statement.AddSource(STATED_IN_SOURCE, PM_ITEM)
				statement.AddSource(RETRIEVED_AT_DATE_SOURCE, fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/11", now.Year(), now.Month(), now.Day()))
				qs_file.WriteString(fmt.Sprintf("%v", statement))
			}

			if license_item != "" {
				statement := AddItemPropertyToItem(item, LICENSE_PROPERTY, license_item)
				statement.AddSource(STATED_IN_SOURCE, license_source)
				statement.AddSource(RETRIEVED_AT_DATE_SOURCE, fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/11", now.Year(), now.Month(), now.Day()))
				qs_file.WriteString(fmt.Sprintf("%v", statement))
			}

			if issn_item != "" {
				statement := AddItemPropertyToItem(item, PUBLICATION_PROPERTY, issn_item)
				statement.AddSource(STATED_IN_SOURCE, PMC_ITEM)
				statement.AddSource(RETRIEVED_AT_DATE_SOURCE, fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/11", now.Year(), now.Month(), now.Day()))
				qs_file.WriteString(fmt.Sprintf("%v", statement))
			}

			for _, subject := range record.MainSubjects {
				if drug_wikidata_items[subject.MeshID] != "" {
					statement = AddItemPropertyToItem(item, MAIN_SUBJECT_PROPERTY, drug_wikidata_items[subject.MeshID])
					statement.AddSource(STATED_IN_SOURCE, PM_ITEM)
					statement.AddSource(RETRIEVED_AT_DATE_SOURCE, fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/11", now.Year(), now.Month(), now.Day()))
					qs_file.WriteString(fmt.Sprintf("%v", statement))
				}
				if disease_wikidata_items[subject.MeshID] != "" {
					statement = AddItemPropertyToItem(item, MAIN_SUBJECT_PROPERTY, disease_wikidata_items[subject.MeshID])
					statement.AddSource(STATED_IN_SOURCE, PM_ITEM)
					statement.AddSource(RETRIEVED_AT_DATE_SOURCE, fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/11", now.Year(), now.Month(), now.Day()))
					qs_file.WriteString(fmt.Sprintf("%v", statement))
				}
			}

			if record.IsRetracted {
				statement := AddItemPropertyToItem(item, INSTANCE_OF_PROPERTY, RETRACTED_PAPER_TYPE)
				statement.AddSource(STATED_IN_SOURCE, PM_ITEM)
				statement.AddSource(RETRIEVED_AT_DATE_SOURCE, fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/11", now.Year(), now.Month(), now.Day()))
				qs_file.WriteString(fmt.Sprintf("%v", statement))

				if retracted_by_item != "" {
					statement := AddItemPropertyToItem(item, RETRACTED_BY_PROPERTY, retracted_by_item)
					statement.AddSource(STATED_IN_SOURCE, PM_ITEM)
					statement.AddSource(RETRIEVED_AT_DATE_SOURCE, fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/11", now.Year(), now.Month(), now.Day()))
					qs_file.WriteString(fmt.Sprintf("%v", statement))
				}
			}

			if record.IsRetraction {
				statement := AddItemPropertyToItem(item, INSTANCE_OF_PROPERTY, RETRACTION_NOTICE_TYPE)
				statement.AddSource(STATED_IN_SOURCE, PM_ITEM)
				statement.AddSource(RETRIEVED_AT_DATE_SOURCE, fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/11", now.Year(), now.Month(), now.Day()))
				qs_file.WriteString(fmt.Sprintf("%v", statement))
			}

			qs_file.WriteString("\n")
		}

		main_subjects := ""
		for idx, subject := range record.MainSubjects {
			if idx != 0 {
				main_subjects += "; "
			}
			main_subjects += subject.Name
			l := drug_wikidata_items[subject.MeshID]
			if disease_wikidata_items[subject.MeshID] != "" {
				if l != "" {
					l += ", "
				}
				l += disease_wikidata_items[subject.MeshID]
			}
			if l != "" {
				main_subjects += fmt.Sprintf(" (%s)", l)
			}
		}

		review_str := "false"
		if record.IsReview {
			review_str = "true"
		}

		retracted_str := "false"
		if record.IsRetracted {
			retracted_str = "true"
		}

		retraction_str := "false"
		if record.IsRetracted {
			retraction_str = "true"
		}

		csv_file.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			record.Title, item, record.PMID, record.PMCID, record.PMCLicense,
			record.EPMCLicenseLink, license_item, main_subjects,
			record.PublicationDate, record.Publication, record.ISSN, issn_item, review_str,
			retracted_str, record.RetractedByPMID, retracted_by_item, retraction_str))
	}

	return nil
}

func main() {

	var term_feed_path string
	var ncbi_api_key string
	flag.StringVar(&term_feed_path, "feed", "", "JSON list of terms to search PMC for.")
	flag.StringVar(&ncbi_api_key, "ncbi_api_key", "", "NCBI API KEY. Can also be set as NCBI_API_KEY environmental variable.")
	flag.Parse()

	if ncbi_api_key == "" {
		ncbi_api_key = os.Getenv("NCBI_API_KEY")
	}

	f, err := os.Open(term_feed_path)
	if err != nil {
		panic(err)
	}

	var term_feed []string
	err = json.NewDecoder(f).Decode(&term_feed)
	if err != nil {
		panic(err)
	}

	qs_file, err := os.Create("results_quickstatements.txt")
	if err != nil {
		panic(err)
	}
	defer qs_file.Close()
	csv_file, err := os.Create("results.csv")
	if err != nil {
		panic(err)
	}
	defer csv_file.Close()
	csv_file.WriteString("Title\tItem\tPMID\tPMCID\tLicense PMC\tLicense EPMC\tLicense Item\tMain Subjects\tPublication Date\tPublication\tISSN\tISSN item\tIs Review Article\tIs retracted\tRetracted by\tRetacted by item\tIs retraction\n")

	for _, term := range term_feed {
		x := fmt.Sprintf("\"%s\"[Mesh Major Topic] AND (Review[ptyp] OR \"Retraction of Publication\"[PTYP])", term)
		err := batch(x, ncbi_api_key, csv_file, qs_file)
		if err != nil {
			panic(err)
		}
	}
}
