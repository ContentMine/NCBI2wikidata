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
	"strings"
)

type Head struct {
	Vars []string `json:"vars"`
}

type Result struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Binding struct {
	Result Result `json:"res"`
	Key    Result `json:"val"`
}

type Results struct {
	Bindings []Binding `json:"bindings"`
}

type SparqlResponse struct {
	Head    Head    `json:"head"`
	Results Results `json:"results"`
}

const SPARQL_QUERY_URL = "https://query.wikidata.org/sparql"

const QUERY_HEADER = `SELECT ?res ?val WHERE {
`
const QUERY_BODY = `
  {
    ?res wdt:P31 wd:%s.
    ?res wdt:%s "%s".
  }
`
const QUERY_FOOTER = `
  OPTIONAL { ?res wdt:%s ?val. }
}
`

func buildSparqlQuery(key string, values []string, item_type string) string {

	query := QUERY_HEADER
	for idx, val := range values {
		if idx != 0 {
			query += " UNION "
		}
		query += fmt.Sprintf(QUERY_BODY, item_type, key, val)
	}
	query += fmt.Sprintf(QUERY_FOOTER, key)

	return query
}

func GetItemsFromWikiData(key string, values []string, item_type string) (map[string]string, error) {

	// If we're not given anything don't bother the server
	results := make(map[string]string)
	if len(values) == 0 {
		return results, nil
	}

	query_string := buildSparqlQuery(key, values, item_type)

	req, err := http.NewRequest("GET", SPARQL_QUERY_URL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("format", "json")
	q.Add("query", query_string)
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

	data := SparqlResponse{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	for _, binding := range data.Results.Bindings {

		// In theory we whouldn't get multiple matches for the things we're looking up
		// (i.e., each PMCID should give us just one paper item back). Due to mistakes that might
		// not be true, so we just log when we hit issues
		val := strings.TrimPrefix(binding.Result.Value, "http://www.wikidata.org/entity/")
		if results[binding.Key.Value] != "" && results[binding.Key.Value] != val {
			results[binding.Key.Value] += "CLASH"
		} else {
			results[binding.Key.Value] = val
		}
	}

	return results, nil
}

func PMCIDsToWDItem(pmcids []string) (map[string]string, error) {
	return GetItemsFromWikiData(PMCID_PROPERTY, pmcids, SCHOLARLY_ARTICLE_TYPE)
}

func PMIDsToWDItem(pmcids []string) (map[string]string, error) {
	return GetItemsFromWikiData(PMID_PROPERTY, pmcids, SCHOLARLY_ARTICLE_TYPE)
}

func ISSNsToWDItem(issn []string) (map[string]string, error) {
	return GetItemsFromWikiData(ISSN_PROPERTY, issn, SCIENTIFIC_JOURNAL_TYPE)
}

func DrugsToWDItem(meshids []string) (map[string]string, error) {
	return GetItemsFromWikiData(MESH_ID_PROPERTY, meshids, DRUG_TYPE)
}

func DiseasesToWDItem(meshids []string) (map[string]string, error) {
	return GetItemsFromWikiData(MESH_ID_PROPERTY, meshids, DISEASE_TYPE)
}
