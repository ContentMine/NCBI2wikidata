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

package EUtils

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const EFETCH_URL string = "https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.fcgi"

type PubmedArticleSet struct {
	XMLName  xml.Name        `xml:"PubmedArticleSet"`
	Articles []PubmedArticle `xml:"PubmedArticle"`
}

type PubmedArticle struct {
	XMLName         xml.Name        `xml:"PubmedArticle"`
	MedlineCitation MedlineCitation `xml:"MedlineCitation"`
	PubMedData      PubMedData      `xml:"PubmedData"`
}

type MedlineCitation struct {
	XMLName                 xml.Name                `xml:"MedlineCitation"`
	Status                  string                  `xml:"Status,attr"`
	Owner                   string                  `xml:"Owner,attr"`
	PMID                    string                  `xml:"PMID"`
	Article                 []Article               `xml:"Article"`
	MeshHeadingList         MeshHeadingList         `xml:"MeshHeadingList"`
	CommentsCorrectionsList CommentsCorrectionsList `xml:"CommentsCorrectionsList"`
}

type Article struct {
	XMLName             xml.Name            `xml:"Article"`
	PubModel            string              `xml:"PubModel,attr"`
	ArticleTitle        string              `xml:"ArticleTitle"`
	Journal             Journal             `xml:"Journal"`
	Language            string              `xml:"Language"`
	PublicationTypeList PublicationTypeList `xml:"PublicationTypeList"`
	ArticleDate         ArticleDate         `xml:"ArticleDate"`
}

type ArticleDate struct {
	XMLName xml.Name `xml:"ArticleDate"`
	Year    int      `xml:"Year"`
	Month   int      `xml:"Month"`
	Day     int      `xml:"Day"`
}

type PublicationTypeList struct {
	XMLName          xml.Name          `xml:"PublicationTypeList"`
	PublicationTypes []PublicationType `xml:"PublicationType"`
}

type PublicationType struct {
	XML  xml.Name `xml:"PublicationType"`
	Type string   `xml:",chardata"`
	UI   string   `xml:"UI,attr"`
}

type Journal struct {
	XMLName         xml.Name     `xml:"Journal"`
	Title           string       `xml:"Title"`
	ISOAbbreviation string       `xml:"ISOAbbreviation"`
	JournalIssue    JournalIssue `xml:"JournalIssue"`
	ISSN            string       `xml:"ISSN"`
}

type JournalIssue struct {
	XMLName    xml.Name `xml:"JournalIssue"`
	CitedMedia string   `xml:"CitedMedia,attr"`
	Volume     string   `xml:"Volume"`
	Issue      string   `xml:"Issue"`
	PubDate    PubDate  `xml:"PubDate"`
}

type PubDate struct {
	XMLName xml.Name `xml:"PubDate"`
	Year    int      `xml:"Year"`
	Month   string   `xml:"Month"`
	Day     int      `xml:"Day"`
}

type MeshHeadingList struct {
	XMLName      xml.Name      `xml:"MeshHeadingList"`
	MeshHeadings []MeshHeading `xml:"MeshHeading"`
}

type MeshHeading struct {
	XMLName        xml.Name            `xml:"MeshHeading"`
	DescriptorName MeshDescriptorName  `xml:"DescriptorName"`
	QualifierNames []MeshQualifierName `xml:"QualifierName"`
}

type MeshDescriptorName struct {
	XMLName      xml.Name `xml:"DescriptorName"`
	Name         string   `xml:",chardata"`
	MeshID       string   `xml:"UI,attr"`
	MajorTopicYN string   `xml:"MajorTopicYN,attr"`
}

type MeshQualifierName struct {
	XMLName      xml.Name `xml:"QualifierName"`
	Name         string   `xml:",chardata"`
	MeshID       string   `xml:"UI,attr"`
	MajorTopicYN string   `xml:"MajorTopicYN,attr"`
}

type CommentsCorrectionsList struct {
	XMLName             xml.Name              `xml:"CommentsCorrectionsList"`
	CommentsCorrections []CommentsCorrections `xml:"CommentsCorrections"`
}

type CommentsCorrections struct {
	XMLName   xml.Name `xml:"CommentsCorrections"`
	RefType   string   `xml:"RefType,attr"`
	RefSource string   `xml:"RefSource"`
	PMID      string   `xml:"PMID"`
}

type PubMedData struct {
	XMLName           xml.Name      `xml:"PubmedData"`
	ArticleIDList     ArticleIdList `xml:"ArticleIdList"`
	PublicationStatus string        `xml:"PublicationStatus"`
}

type ArticleIdList struct {
	XMLName    xml.Name    `xml:"ArticleIdList"`
	ArticleIDs []ArticleID `xml:"ArticleId"`
}

type ArticleID struct {
	XMLName xml.Name `xml:"ArticleId"`
	ID      string   `xml:",chardata"`
	IDType  string   `xml:"IdType,attr"`
}

type EFetchHistoryRequest struct {
	DB       string
	WebEnv   string
	QueryKey string
	APIKey   string
	RetMax   int
	RetStart int
}

func (e *EFetchHistoryRequest) Do() (PubmedArticleSet, error) {

	if e.APIKey == "" {
		return PubmedArticleSet{}, fmt.Errorf("No API Key provided.")
	}

	req, err := http.NewRequest("GET", EFETCH_URL, nil)
	if err != nil {
		return PubmedArticleSet{}, err
	}

	q := req.URL.Query()
	q.Add("api_key", e.APIKey)
	q.Add("db", e.DB)
	q.Add("WebEnv", e.WebEnv)
	q.Add("query_key", e.QueryKey)
	q.Add("retmode", "xml")
	if e.RetMax > 0 {
		q.Add("retmax", fmt.Sprintf("%d", e.RetMax))
	}
	if e.RetStart > 0 {
		q.Add("retstart", fmt.Sprintf("%d", e.RetStart))
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return PubmedArticleSet{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return PubmedArticleSet{}, fmt.Errorf("Status code %d", resp.StatusCode)
		} else {
			return PubmedArticleSet{}, fmt.Errorf("Status code %d: %s", resp.StatusCode, body)
		}
	}

	efetch_resp := PubmedArticleSet{}
	err = xml.NewDecoder(resp.Body).Decode(&efetch_resp)
	if err != nil {
		return PubmedArticleSet{}, err
	}

	return efetch_resp, nil
}

func (article PubmedArticle) GetPMID() string {
	return article.MedlineCitation.PMID
}

func (article PubmedArticle) GetPMCID() string {

	for _, articleID := range article.PubMedData.ArticleIDList.ArticleIDs {
		if articleID.IDType == "pmc" {
			return strings.TrimPrefix(articleID.ID, "PMC")
		}
	}
	return ""
}

func (article PubmedArticle) GetMajorTopics() []MeshDescriptorName {

	subjects := make([]MeshDescriptorName, 0)
	for _, mesh := range article.MedlineCitation.MeshHeadingList.MeshHeadings {
		major := mesh.DescriptorName.MajorTopicYN == "Y"
		for _, qual := range mesh.QualifierNames {
			major = major || qual.MajorTopicYN == "Y"
		}
		if major {
			subjects = append(subjects, mesh.DescriptorName)
		}
	}

	return subjects
}

var MONTH_TO_INT = map[string]int{
	"jan": 1,
	"feb": 2,
	"mar": 3,
	"apr": 4,
	"may": 5,
	"jun": 6,
	"jul": 7,
	"aug": 8,
	"sep": 9,
	"oct": 10,
	"nov": 11,
	"dec": 12,
}

func monthStringToInt(m string) (int, error) {

	// Sometimes it's a number as a string
	x, err := strconv.Atoi(m)
	if err == nil {
		return x, nil
	}

	// sometimes it's as human readable shorted text
	v, ok := MONTH_TO_INT[strings.ToLower(m)]
	if ok {
		return v, nil
	}

	return 0, fmt.Errorf("Failed to translate month %s to number", m)
}

func (article PubmedArticle) GetPublicationDateString() string {

	pubdate := ""
	p := article.MedlineCitation.Article[0].Journal.JournalIssue.PubDate
	if p.Month != "" && p.Year != 0 {
		pmonth, err := monthStringToInt(p.Month)
		if err == nil && p.Year != 0 {
			if p.Day == 0 {
				p.Day = 1
			}
			pubdate = fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/11", p.Year, pmonth, p.Day)
		}
	}
	if pubdate == "" {
		p := article.MedlineCitation.Article[0].ArticleDate
		if p.Month != 0 && p.Year != 0 {
			if p.Day == 0 {
				p.Day = 1
			}
			pubdate = fmt.Sprintf("+%04d-%02d-%02dT00:00:00Z/11", p.Year, p.Month, p.Day)
		}
	}

	return pubdate
}

func (article PubmedArticle) IsReview() bool {

	for _, pubtype := range article.MedlineCitation.Article[0].PublicationTypeList.PublicationTypes {
		if pubtype.Type == "Review" || pubtype.Type == "Systematic Review" {
			return true
		}
	}
	return false
}

func (article PubmedArticle) IsRetracted() bool {

	for _, pubtype := range article.MedlineCitation.Article[0].PublicationTypeList.PublicationTypes {
		if pubtype.Type == "Retracted Publication" {
			return true
		}
	}
	return false
}

func (article PubmedArticle) IsRetraction() bool {

	for _, pubtype := range article.MedlineCitation.Article[0].PublicationTypeList.PublicationTypes {
		if pubtype.Type == "Retraction of Publication" {
			return true
		}
	}
	return false
}

func (article PubmedArticle) GetRetractedInPMID() string {

	for _, comment := range article.MedlineCitation.CommentsCorrectionsList.CommentsCorrections {
		if comment.RefType == "RetractionIn" {
			return comment.PMID
		}
	}
	return ""
}
