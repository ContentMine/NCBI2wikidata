#!/usr/bin/env python
#
#   Copyright 2019 Content Mine Ltd
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

import json
import requests

TOP_LEVEL = """
SELECT ?spec ?specLabel ?count
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
"""

REFINE_QUERY = """
SELECT DISTINCT ?item ?MeSHID ?itemLabel
  WHERE {
  ?item wdt:P31 wd:Q12136;
        wdt:P1995 ?medspec;
        wdt:P486 ?MeSHID.
  ?medspec wdt:P361* wd:%s .

  SERVICE wikibase:label { bd:serviceParam wikibase:language "en" }
}
"""



r = requests.post("https://query.wikidata.org/sparql", {"query": TOP_LEVEL, "format": "json"})

list = []

for c in r.json()["results"]["bindings"]:

    item_url = c['spec']['value']
    item_code = item_url.split('/')[-1]

    r = requests.post("https://query.wikidata.org/sparql", {"query": REFINE_QUERY % item_code, "format": "json"})

    results = r.json()["results"]["bindings"]
    print "We got %d specifics for %s" % (len(results), c['specLabel']['value'])

    for d in results:
        list.append(d["itemLabel"]["value"])


print "We found %d terms" % len(list)

with open("generated_feed.json", "w") as f:
    json.dump(list, f, indent=4)
