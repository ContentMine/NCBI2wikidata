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
	"os"
	"testing"
)

func loadXML(filename string) (PubmedArticleSet, error) {
	var set PubmedArticleSet

	f, err := os.Open(filename)
	if err != nil {
		return PubmedArticleSet{}, err
	}

	err = xml.NewDecoder(f).Decode(&set)
	return set, err
}

func TestGenericExample(t *testing.T) {
	article_set, err := loadXML("testdata/example1.xml")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	if len(article_set.Articles) != 1 {
		t.Fatalf("Unexpected number of articles: %d", len(article_set.Articles))
	}

	article := article_set.Articles[0]

	pmid := article.GetPMID()
	if pmid != "29846473" {
		t.Errorf("Got unexpected PMID for article: %s", pmid)
	}

	pmcid := article.GetPMCID()
	if pmcid != "5975557" {
		t.Errorf("Got unexpected PMCID for article: %s", pmcid)
	}

	subjects := article.GetMajorTopics()
	if len(subjects) != 6 {
		t.Errorf("Wrong number of major topics: %d", len(subjects))
	}

	if !article.IsReview() {
		t.Errorf("Expected article to be review.")
	}

	if article.IsRetracted() {
		t.Errorf("Expected article to not be retracted")
	}

	if article.IsRetraction() {
		t.Errorf("Expected article to not be a retraction")
	}

	retracted_in := article.GetRetractedInPMID()
	if retracted_in != "" {
		t.Errorf("Got unexpected retraction PMID: %s", retracted_in)
	}
}

func TestDescriptionOnlyMajorTopics(t *testing.T) {
	article_set, err := loadXML("testdata/topics.xml")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	if len(article_set.Articles) != 1 {
		t.Fatalf("Unexpected number of articles: %d", len(article_set.Articles))
	}

	article := article_set.Articles[0]

	// This paper only gets descriptions as major topics, not qualifiers
	subjects := article.GetMajorTopics()
	if len(subjects) != 2 {
		t.Errorf("Wrong number of major topics: %d", len(subjects))
	}
}

func TestRetractedArticle(t *testing.T) {
	article_set, err := loadXML("testdata/retracted.xml")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	if len(article_set.Articles) != 1 {
		t.Fatalf("Unexpected number of articles: %d", len(article_set.Articles))
	}

	article := article_set.Articles[0]

	if !article.IsRetracted() {
		t.Errorf("Expected article to be retracted")
	}

	if article.IsRetraction() {
		t.Errorf("Expected article to not be a retraction")
	}

	retracted_in := article.GetRetractedInPMID()
	if retracted_in != "30683838" {
		t.Errorf("Got unexpected retraction PMID: %s", retracted_in)
	}
}

func TestRetraction(t *testing.T) {
	article_set, err := loadXML("testdata/retraction.xml")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	if len(article_set.Articles) != 1 {
		t.Fatalf("Unexpected number of articles: %d", len(article_set.Articles))
	}

	article := article_set.Articles[0]

	if article.IsRetracted() {
		t.Errorf("Expected article to be retracted")
	}

	if !article.IsRetraction() {
		t.Errorf("Expected article to not be a retraction")
	}

	retracted_in := article.GetRetractedInPMID()
	if retracted_in != "" {
		t.Errorf("Got unexpected retraction PMID: %s", retracted_in)
	}
}
