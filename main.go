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
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var CC_LICENSE_ITEM_IDS = map[string]string{
	"CC0":         "Q6938433",
	"CC BY":       "Q6905323",
	"CC BY-NC-ND": "Q6937225",
	"CC BY-NC":    "Q6936496",

	// These aren't in the NCBI OA list, but we might get them later from
	// the EuroPMC API
	"CC BY 2.5": "Q18810333",
	"CC BY 4.0": "Q20007257",
}

const EFETCH_BATCH_SIZE = 200

// Load the NCBI open access file list so we can map PMID -> Copyright
//
// First line is date file was generated, rest are tab separated info on papers. Example:
// oa_package/87/30/PMC17774.tar.gz	Arthritis Res. 1999 Oct 14; 1(1):63-70	PMC17774	PMID:11056661	NO-CC CODE
//
func LoadLicenses(filename string) (map[string]string, error) {

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	lookup := make(map[string]string)

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

		pmid := parts[3]
		license := strings.TrimSuffix(parts[4], "\n")

		split_pmid := strings.Split(pmid, ":")
		if len(split_pmid) == 2 {
			// Not all entries have a PMID, but this one did
			lookup[split_pmid[1]] = license
		}

		// Also add in the PMCID
		if len(parts[2]) > 0 {
			lookup[parts[2]] = license
		} else {
			fmt.Printf("%s", line)
		}
	}

	return lookup, nil
}

type Record struct {
	Title           string
	MainSubjects    []MeshDescriptorName
	PublicationType string
	PublicationDate string
	Publication     string
	ISSN            string
	License         string
	PMID            string
	PMCID           string
}

func batch(term string) error {

	license_lookup, err := LoadLicenses("oa_file_list.txt")
	if err != nil {
		return err
	}


    // Because we use the history feature of the eUtilities API, it doesn't matter how many
    // things get returned here, we rely on the eFetch API to get all the deets. Hence the
    // single request here. We really are just doing this to light up things later
	search_request := ESearchRequest{
		DB:         "pubmed",
		APIKey:     os.Getenv("NCBI_API_KEY"),
		Term:       term,
		RetMax:     1,
		UseHistory: true,
	}

	search_resp, rerr := search_request.Do()
	if rerr != nil {
		return rerr
	}

    count, count_err := strconv.Atoi(search_resp.Count)
    if count_err != nil {
        return count_err
    }
	log.Printf("Search returned %d matches for %s.\n", count, term)

    // Things to build up as we fetch the results from PMC...
	records := make([]Record, 0)
	pmcid_list := make([]string, 0)
	issn_list := make([]string, 0)
	main_subject_list := make([]string, 0)

	for i := 0; i < count; i += EFETCH_BATCH_SIZE {

        fetch_request := EFetchHistoryRequest{
            DB:       "pubmed",
            WebEnv:   search_resp.WebEnv,
            QueryKey: search_resp.QueryKey,
            APIKey:   os.Getenv("NCBI_API_KEY"),
            RetStart: i,
            RetMax:   EFETCH_BATCH_SIZE,
        }
        time.Sleep(100 * time.Millisecond)

        fetch_resp, ferr := fetch_request.Do()
        if ferr != nil {
            return ferr
        }

        log.Printf("Fetched %d articles for %s.\n", len(fetch_resp.Articles), term)


        for _, article := range fetch_resp.Articles {

            // Is there a PMCID for this paper
            var pmcid string
            for _, articleID := range article.PubMedData.ArticleIDList.ArticleIDs {
                if articleID.IDType == "pmc" {
                    pmcid = strings.TrimPrefix(articleID.ID, "PMC")
                    break
                }
            }
            if pmcid != "" {
                pmcid_list = append(pmcid_list, pmcid)
            }

            var license string
            if l, ok := license_lookup[article.MedlineCitation.PMID]; ok {
                license = l
            }
            if license == "" {
                // didn't find one with PMID, try PMCID
                if l, ok := license_lookup[pmcid]; ok {
                    license = l
                }
            }

            if license == "" {
                continue
            }

            subjects := make([]MeshDescriptorName, 0)
            for _, mesh := range article.MedlineCitation.MeshHeadingList.MeshHeadings {
                major := mesh.DescriptorName.MajorTopicYN == "Y"
                for _, qual := range mesh.QualifierNames {
                    major = major || qual.MajorTopicYN == "Y"
                }
                if major {
                    main_subject_list = append(main_subject_list, mesh.DescriptorName.MeshID)
                    subjects = append(subjects, mesh.DescriptorName)
                }
            }

            p := article.MedlineCitation.Article[0].Journal.JournalIssue.PubDate
            pubdate := fmt.Sprintf("%s-%d", p.Month, p.Year)
            if p.Year == 0 || p.Month == "" {
                p := article.MedlineCitation.Article[0].ArticleDate
                pubdate = fmt.Sprintf("%d-%d", p.Month, p.Year)
            }

            var pubtype string
            if len(article.MedlineCitation.Article[0].PublicationTypeList.PublicationTypes) > 0 {
                pubtype = article.MedlineCitation.Article[0].PublicationTypeList.PublicationTypes[0].Type
            }

            issn := article.MedlineCitation.Article[0].Journal.ISSN
            if issn != "" {
                issn_list = append(issn_list, issn)
            }

            r := Record{
                Title:           article.MedlineCitation.Article[0].ArticleTitle,
                PMID:            article.MedlineCitation.PMID,
                PMCID:           pmcid,
                License:         license,
                MainSubjects:    subjects,
                PublicationDate: pubdate,
                Publication:     article.MedlineCitation.Article[0].Journal.Title,
                ISSN:            issn,
                PublicationType: pubtype,
            }

            records = append(records, r)
        }
    }

	log.Printf("We got information on %d records.\n", len(records))

	pmcid_wikidata_items, perr := PMCIDsToWDItem(pmcid_list)
	if perr != nil {
		return fmt.Errorf("Failed fetching %d PMCID items: %v", len(pmcid_list), perr)
	}
	issn_wikidata_items, ierr := ISSNsToWDItem(issn_list)
	if ierr != nil {
		return fmt.Errorf("Failed fetching %d ISSN items: %v", len(issn_list), ierr)
	}
	drug_wikidata_items, d1err := DrugsToWDItem(main_subject_list)
	if d1err != nil {
		return fmt.Errorf("Failed fetching drug %d items: %v", len(main_subject_list), d1err)
	}
	disease_wikidata_items, d2err := DiseasesToWDItem(main_subject_list)
	if d2err != nil {
		return fmt.Errorf("Failed fetching %d disease items: %v", len(main_subject_list), d2err)
	}

	qs_file, qe := os.Create("results_quickstatements.txt")
	if qe != nil {
		return qe
	}
	defer qs_file.Close()
	csv_file, ce := os.Create("results.csv")
	if ce != nil {
		return ce
	}
	defer csv_file.Close()
	csv_file.WriteString("Title\tItem\tPMID\tPMCID\tLicense\tLicense Item\tMain Subjects\tPublication Date\tPublication\tISSN\tISSN item\tPublication Type\n")

	now := time.Now()

	for _, record := range records {

		item := pmcid_wikidata_items[record.PMCID]
		issn_item := issn_wikidata_items[record.ISSN]

		if item != "" {
			statement := AddStringPropertyToItem(item, PMID_PROPERTY, record.PMID)
			statement.AddSource(STATED_IN_SOURCE, PMC_ITEM)
			statement.AddSource(RETRIEVED_AT_DATE_SOURCE, fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/9", now.Year(), now.Month(), now.Day()))
			qs_file.WriteString(fmt.Sprintf("%v", statement))

			statement = AddStringPropertyToItem(item, PMCID_PROPERTY, record.PMCID)
			statement.AddSource(STATED_IN_SOURCE, PMC_ITEM)
			statement.AddSource(RETRIEVED_AT_DATE_SOURCE, fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/9", now.Year(), now.Month(), now.Day()))
			qs_file.WriteString(fmt.Sprintf("%v", statement))

			if CC_LICENSE_ITEM_IDS[record.License] != "" {
				statement = AddItemPropertyToItem(item, LICENSE_PROPERTY, CC_LICENSE_ITEM_IDS[record.License])
				statement.AddSource(STATED_IN_SOURCE, PMC_ITEM)
				statement.AddSource(RETRIEVED_AT_DATE_SOURCE, fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/9", now.Year(), now.Month(), now.Day()))
				qs_file.WriteString(fmt.Sprintf("%v", statement))
			}

			if issn_item != "" {
				statement = AddItemPropertyToItem(item, PUBLICATION_PROPERTY, issn_item)
				statement.AddSource(STATED_IN_SOURCE, PMC_ITEM)
				statement.AddSource(RETRIEVED_AT_DATE_SOURCE, fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/9", now.Year(), now.Month(), now.Day()))
				qs_file.WriteString(fmt.Sprintf("%v", statement))
			}

			for _, subject := range record.MainSubjects {
				if drug_wikidata_items[subject.MeshID] != "" {
					statement = AddItemPropertyToItem(item, MAIN_SUBJECT_PROPERTY, drug_wikidata_items[subject.MeshID])
					statement.AddSource(STATED_IN_SOURCE, PMC_ITEM)
					statement.AddSource(RETRIEVED_AT_DATE_SOURCE, fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/9", now.Year(), now.Month(), now.Day()))
					qs_file.WriteString(fmt.Sprintf("%v", statement))
				}
				if disease_wikidata_items[subject.MeshID] != "" {
					statement = AddItemPropertyToItem(item, MAIN_SUBJECT_PROPERTY, disease_wikidata_items[subject.MeshID])
					statement.AddSource(STATED_IN_SOURCE, PMC_ITEM)
					statement.AddSource(RETRIEVED_AT_DATE_SOURCE, fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/9", now.Year(), now.Month(), now.Day()))
					qs_file.WriteString(fmt.Sprintf("%v", statement))
				}
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

		csv_file.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			record.Title, item, record.PMID, record.PMCID, record.License,
			CC_LICENSE_ITEM_IDS[record.License], main_subjects,
			record.PublicationDate, record.Publication, record.ISSN, issn_item, record.PublicationType))
	}

	return nil
}

func main() {

	var term_feed_path string
	flag.StringVar(&term_feed_path, "feed", "", "JSON list of terms to search PMC for.")
	flag.Parse()

	f, err := os.Open(term_feed_path)
	if err != nil {
		panic(err)
	}

	var term_feed []string
	err = json.NewDecoder(f).Decode(&term_feed)
	if err != nil {
		panic(err)
	}

	for _, term := range term_feed[:1] {
		x := fmt.Sprintf("\"%s\"[Mesh Major Topic] AND Review[ptyp]", term)
		err := batch(x)
		if err != nil {
			panic(err)
		}
	}
}
