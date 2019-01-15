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
	"encoding/xml"
	//	"fmt"
	//	"net/http"
)

const EFETCH_URL string = "https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.fcgi"

type PubmedArticleSet struct {
	XMLName  xml.Name        `xml:"PubmedArticleSet"`
	Articles []PubmedArticle `xml:"PubmedArticle"`
}

type PubmedArticle struct {
	XMLName         xml.Name        `xml:"PubmedArticle"`
	MedlineCitation MedlineCitation `xml:"MedlineCitation"`
	PubMedData      PubMedData      `xml:"PubMedData"`
}

type MedlineCitation struct {
	XMLName         xml.Name        `xml:"MedlineCitation"`
	Status          string          `xml:"Status,attr"`
	Owner           string          `xml:"Owner,attr"`
	PMID            string          `xml:"PMID"`
	Article         []Article       `xml:"Article"`
	MeshHeadingList MeshHeadingList `xml:"MeshHeadingList"`
}

type Article struct {
	XMLName      xml.Name `xml:"Article"`
	PubModel     string   `xml:"PubModel,attr"`
	ArticleTitle string   `xml:"ArticleTitle"`
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
	UI           string   `xml:"UI,attr"`
	MajorTopicYN string   `xml:MajorTopicYN,attr"`
}

type MeshQualifierName struct {
	XMLName      xml.Name `xml:"QualifierName"`
	Name         string   `xml:",chardata"`
	UI           string   `xml:"UI,attr"`
	MajorTopicYN string   `xml:MajorTopicYN,attr"`
}

type PubMedData struct {
	XMLName       xml.Name    `xml:"PubMedData"`
	ArticleIDList []ArticleID `xml:"ArticleIdList"`
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
}
