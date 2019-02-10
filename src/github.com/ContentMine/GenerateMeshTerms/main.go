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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

type MeSHLabel struct {
	Language string `json:"@language"`
	Value    string `json:"@value"`
}

type MeSHGraph struct {
	Identifier string    `json:"identifier"`
	Label      MeSHLabel `json:"label"`
}

type MeSHIDLookup struct {
	MeSHGraph []MeSHGraph `json:"@graph"`
}

type Head struct {
	Vars []string `json:"vars"`
}

type Result struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// This is a superset of all the queries we might make, to save us playing clever with Go.
type Binding struct {
	Spec   Result `json:"spec"`
	MeSHID Result `json:"MeSHID"`
}

type Results struct {
	Bindings []Binding `json:"bindings"`
}

type SparqlResponse struct {
	Head    Head    `json:"head"`
	Results Results `json:"results"`
}

const TOP_LEVEL = `
SELECT ?spec ?specLabel
WHERE
{
  {
    SELECT ?spec (COUNT(?item) AS ?count)
WHERE {
        ?item wdt:P31 wd:Q12136 .
        ?item wdt:P1995 ?spec  .
        }
  GROUP BY ?spec
  }
   SERVICE wikibase:label { bd:serviceParam wikibase:language "en" }
}
`

const REFINE_QUERY = `
SELECT DISTINCT ?item ?MeSHID ?itemLabel
  WHERE {
  ?item wdt:P31 wd:Q12136;
        wdt:P1995 ?medspec;
        wdt:P486 ?MeSHID.
  ?medspec wdt:P361* wd:%s .

  SERVICE wikibase:label { bd:serviceParam wikibase:language "en" }
}
`

const SPARQL_QUERY_URL = "https://query.wikidata.org/sparql"

const MESH_LABEL_URL = "https://id.nlm.nih.gov/mesh/%s.json"

func makeWikidataQuery(query string) ([]Binding, error) {

	params := url.Values{}
	params.Add("query", query)

	req, err := http.NewRequest("POST", SPARQL_QUERY_URL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/sparql-results+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
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

	return data.Results.Bindings, nil
}

func getMeshLabel(mesh_id string) (string, error) {

	if mesh_id == "" {
		return "", fmt.Errorf("Empty MeSH ID provided")
	}

	mesh_url := fmt.Sprintf(MESH_LABEL_URL, mesh_id)

	req, err := http.NewRequest("GET", mesh_url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("Status code %d", resp.StatusCode)
		} else {
			return "", fmt.Errorf("Status code %d: %s", resp.StatusCode, body)
		}
	}

	data := MeSHIDLookup{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	if len(data.MeSHGraph) != 1 {
		return "", fmt.Errorf("We got %d results for %s", len(data.MeSHGraph), mesh_id)
	}

	return data.MeSHGraph[0].Label.Value, nil
}

func main() {

	var ncbi_api_key string
	flag.StringVar(&ncbi_api_key, "ncbi_api_key", "", "NCBI API KEY. Can also be set as NCBI_API_KEY environmental variable.")
	flag.Parse()

	if ncbi_api_key == "" {
		ncbi_api_key = os.Getenv("NCBI_API_KEY")
	}

	specialities, err := makeWikidataQuery(TOP_LEVEL)
	if err != nil {
		panic(err)
	}

	log.Printf("speciaily count: %d", len(specialities))

	meshid_set := make(map[string]string, 0)

	for _, binding := range specialities {
		if binding.Spec.Value == "" {
			panic(fmt.Errorf("We got an empty spec binding: %v", binding))
		}

		query := fmt.Sprintf(REFINE_QUERY, strings.TrimPrefix(binding.Spec.Value, "http://www.wikidata.org/entity/"))
		specifics, err := makeWikidataQuery(query)
		if err != nil {
			panic(err)
		}

		for _, specific_binding := range specifics {
			if specific_binding.MeSHID.Value == "" {
				panic(fmt.Errorf("We got an empty specific binding: %v", binding))
			}
			meshid_set[specific_binding.MeSHID.Value] = ""
		}
	}

	log.Printf("MeSH ID count: %d", len(meshid_set))

	for mesh_id, _ := range meshid_set {

		if mesh_id == "NoID" {
			continue
		}

		label, err := getMeshLabel(mesh_id)
		if err != nil {
			log.Printf("Error looking up %s, skipping: %v.", mesh_id, err)
			continue
		}

		meshid_set[mesh_id] = label

		// The NCBI docs ask in general that you don't make request more frequently than ten a second. I'm not
		// sure that this end point is covered under that (as I suspect this is a static page) but just incase...
		time.Sleep(500 * time.Millisecond)
	}

	label_list := make([]string, 0, len(meshid_set))
	for _, label := range meshid_set {
		if label == "" {
			continue
		}
		label_list = append(label_list, label)
	}
	sort.Strings(label_list)

	feed, err := os.Create("generated_feed.json")
	if err != nil {
		panic(err)
	}
	defer feed.Close()

	err = json.NewEncoder(feed).Encode(label_list)
	if err != nil {
		panic(err)
	}

}
