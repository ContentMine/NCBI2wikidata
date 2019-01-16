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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const ESEARCH_URL string = "https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi"

// ?db=pubmed&term=food[MeSH%20Major%20Topic]&reldate=60&datetype=edat&retmax=100&usehistory=y&retmode=json

type ESearchHeader struct {
	Type    string `json:"type"`
	Version string `json:"version"`
}

type ESearchTranslationSetItem struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ESearchTranslationStackItem struct {
	Term    string `json:"term"`
	Field   string `json:"field"`
	Count   string `json:"count"`
	Explode string `json:"explode"`
}

type ESearchResult struct {
	Error            *string                       `json:"ERROR"`
	Count            string                        `json:"count"`
	Maximum          string                        `json:"retmax"`
	Start            string                        `json:"restart"`
	QueryKey         string                        `json:"queryKey"`
	WebEnv           string                        `json:"webenv"`
	IDs              []string                      `json:"idlist"`
	TranslationSet   []ESearchTranslationSetItem   `json:"translationset"`
	TranslationStack []ESearchTranslationStackItem `json:"translationstrack"`
	QueryTranslation string                        `json:"querytranslation"`
}

type ESearchResponse struct {
	Header *ESearchHeader `json:"header"`
	Result *ESearchResult `json:"esearchresult"`

	Error *string `json:"error"`
	Count *string `json:"count"`
}

type ESearchRequest struct {
	DB         string
	APIKey     string
	Term       string
	RetMax     int
	RetStart   int
	UseHistory bool
}

func (r ESearchResponse) String() string {

	if r.Error != nil {
		count := "unknown"
		if r.Count != nil {
			count = *r.Count
		}
		return fmt.Sprintf("<Error: %s\tCount: %s>", *r.Error, count)
	} else if r.Result != nil {
		return fmt.Sprintf("<Term: %s\nCount: %s\tReturned: %d>", r.Result.QueryTranslation, r.Result.Count, len(r.Result.IDs))
	} else {
		return fmt.Sprintf("<Empty esearch reply>")
	}
}

func (e *ESearchRequest) Do() (*ESearchResult, error) {

	if e.APIKey == "" {
		return nil, fmt.Errorf("No API Key provided.")
	}

	req, err := http.NewRequest("GET", ESEARCH_URL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("api_key", e.APIKey)
	q.Add("term", e.Term)
	q.Add("db", e.DB)
	q.Add("retmode", "json")
	if e.UseHistory {
		q.Add("usehistory", "y")
	}
	if e.RetMax > 0 {
		q.Add("retmax", fmt.Sprintf("%d", e.RetMax))
	}
	if e.RetStart > 0 {
		q.Add("retstart", fmt.Sprintf("%d", e.RetStart))
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, response_err := client.Do(req)
	if response_err != nil {
		return nil, response_err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("Status code %d", resp.StatusCode)
		} else {
			return nil, fmt.Errorf("Status code %d: %s", resp.StatusCode, body)
		}
	}

	esearch_resp := ESearchResponse{}
	err = json.NewDecoder(resp.Body).Decode(&esearch_resp)
	if err != nil {
		return nil, err
	}

	if esearch_resp.Error != nil {
		return nil, fmt.Errorf("API Error: %s", *esearch_resp.Error)
	}
	if esearch_resp.Result == nil {
		return nil, fmt.Errorf("API Error: No result returned %v", esearch_resp)
	}
	if esearch_resp.Result.Error != nil {
		return nil, fmt.Errorf("Search error: %s", *(esearch_resp.Result.Error))
	}

	return esearch_resp.Result, nil
}
