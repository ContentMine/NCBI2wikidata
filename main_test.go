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
	"testing"
)

func TestRercordGen(t *testing.T) {

	testdata := []struct {
		filename         string
		PMID             string
		PMCID            string
		MainSubjectCount int
		IsReview         bool
		IsRetracted      bool
		IsRetraction     bool
		RetractedByPMID  string
	}{
		{
			filename:         "testdata/example1.xml",
			PMID:             "29846473",
			PMCID:            "5975557",
			MainSubjectCount: 6,
			IsReview:         true,
			IsRetracted:      false,
			IsRetraction:     false,
			RetractedByPMID:  "",
		},
		{
			filename:         "testdata/topics.xml",
			PMID:             "28405850",
			PMCID:            "5486469",
			MainSubjectCount: 2,
			IsReview:         true,
			IsRetracted:      false,
			IsRetraction:     false,
			RetractedByPMID:  "",
		},
		{
			filename:         "testdata/retracted.xml",
			PMID:             "27685632",
			PMCID:            "5059863",
			MainSubjectCount: 4,
			IsReview:         false,
			IsRetracted:      true,
			IsRetraction:     false,
			RetractedByPMID:  "30683838",
		},
		{
			filename:         "testdata/retraction.xml",
			PMID:             "30683838",
			PMCID:            "6347590",
			MainSubjectCount: 0,
			IsReview:         false,
			IsRetracted:      false,
			IsRetraction:     true,
			RetractedByPMID:  "",
		},
	}

	for _, testitem := range testdata {

		article_set, err := loadXML(testitem.filename)
		if err != nil {
			t.Errorf("Failed to load test data: %v", err)
			continue
		}

		if len(article_set.Articles) != 1 {
			t.Errorf("Unexpected number of articles: %d", len(article_set.Articles))
			continue
		}

		article := article_set.Articles[0]

		license_lookup := make(map[string]string, 0)
		license_lookup[article.GetPMID()] = "TEST"

		record, err := ArticleToRecord(article, license_lookup)

		if err != nil {
			t.Fatalf("Failed to process article to record: %v", err)
		}

        if record.PMCLicense != "TEST" {
			t.Errorf("License in record incorrect: %s not %s", record.PMCLicense, "TEST")
        }

		if record.PMID != testitem.PMID {
			t.Errorf("PMID in record incorrect: %s not %s", record.PMID, testitem.PMID)
		}
		if record.PMCID != testitem.PMCID {
			t.Errorf("PMCID in record incorrect: %s not %s", record.PMCID, testitem.PMCID)
		}
		if len(record.MainSubjects) != testitem.MainSubjectCount {
			t.Errorf("Subject count in record incorrect: %d not %d", len(record.MainSubjects), testitem.MainSubjectCount)
		}
	}
}