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
    Type string `json:"type"`
    Value string `json:"value"`
}

type Binding struct {
    Result Result `json:"x"`
}

type Results struct {
    Bindings []Binding `json:"bindings"`
}

type SparqlResponse struct {
    Head Head `json:"head"`
    Results Results `json:"results"`
}


const SPARQL_QUERY_URL = "https://query.wikidata.org/sparql"

const CC_LICENSE_TYPE = "Q284742"
const SCHOLARLY_ARTICLE_TYPE = "Q13442814"
const SCIENTIFIC_JOURNAL_TYPE = "Q5633421"

const PMCID_PROPERTY = "P932"
const ISSN_PROPERTY = "P236"

const QUERY_FORMAT = `SELECT ?x WHERE {
  SERVICE wikibase:label { bd:serviceParam wikibase:language "[AUTO_LANGUAGE],en". }
  ?x wdt:P31 wd:%s.
  ?x wdt:%s "%s".
}`


func GetItemsFromWikiData(key string, value string, item_type string) ([]string, error) {

    req, err := http.NewRequest("GET", SPARQL_QUERY_URL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("format", "json")
	q.Add("query", fmt.Sprintf(QUERY_FORMAT, item_type, key, value))
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

	results := make([]string, len(data.Results.Bindings))
	for idx, binding := range data.Results.Bindings {
	    results[idx] = strings.TrimPrefix(binding.Result.Value, "http://www.wikidata.org/entity/")
	}

	return results, nil
}



func PMCIDToWDItem(pmcid string) (string, error) {
    results, err := GetItemsFromWikiData(PMCID_PROPERTY, strings.TrimPrefix(pmcid, "PMC"), SCHOLARLY_ARTICLE_TYPE)
    if err != nil {
        return "", err
    }
    if len(results) != 1 {
        return "", fmt.Errorf("We wanted just one result, we got %d", len(results))
    } else {
        return results[0], nil
    }
}



func ISSNToWDItem(issn string) (string, error) {
    results, err := GetItemsFromWikiData(ISSN_PROPERTY, issn, SCIENTIFIC_JOURNAL_TYPE)
    if err != nil {
        return "", err
    }
    if len(results) != 1 {
        return "", fmt.Errorf("We wanted just one result, we got %d", len(results))
    } else {
        return results[0], nil
    }
}
