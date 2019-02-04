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

// These items are used in property P31 "instance of"
const CC_LICENSE_TYPE = "Q284742"
const SCHOLARLY_ARTICLE_TYPE = "Q13442814"
const SCIENTIFIC_JOURNAL_TYPE = "Q5633421"
const RETRACTED_PAPER_TYPE = "Q45182324"
const RETRACTION_NOTICE_TYPE = "Q7316896"
const DISEASE_TYPE = "Q12136"
const DRUG_TYPE = "Q8386"

const INSTANCE_OF_PROPERTY = "P31"
const ISSN_PROPERTY = "P236"
const LICENSE_PROPERTY = "P275"
const MAIN_SUBJECT_PROPERTY = "P921"
const MESH_ID_PROPERTY = "P486"
const PMID_PROPERTY = "P698"
const PMCID_PROPERTY = "P932"
const PUBLICATION_PROPERTY = "P1433"
const PUBLICATION_DATE_PROPERTY = "P577"
const TITLE_PROPERTY = "P1476"
const RETRACTED_BY_PROPERTY = "P5824"

const OFFICIAL_WEBSITE_SOURCE = "S856"
const STATED_IN_SOURCE = "S248"
const REFERENCE_URL_SOURCE = "S854"
const RETRIEVED_AT_DATE_SOURCE = "S813"

const PMC_ITEM = "Q229883"
const EuroPMC_ITEM = "Q5412157"
const REVIEW_ARTICLE_ITEM = "Q7318358"

var CC_LICENSE_ITEM_IDS = map[string]string{
	"CC0":         "Q6938433",
	"CC BY":       "Q6905323",
	"CC BY-NC-ND": "Q6937225",
	"CC BY-NC":    "Q6936496",

	// These aren't in the NCBI OA list, but we might get them later from
	// the EuroPMC API
	"https://creativecommons.org/publicdomain/zero/1.0/":        "Q6938433",
	"https://creativecommons.org/publicdomain/mark/1.0/":        "Q7257361",
	"https://creativecommons.org/licenses/by-sa/3.0/":           "Q14946043",
	"https://creativecommons.org/licenses/by/3.0/":              "Q14947546",
	"https://creativecommons.org/licenses/by-nc-sa/3.0/":        "Q15643954",
	"https://creativecommons.org/licenses/by-sa/2.5/se":         "Q15914252",
	"https://creativecommons.org/licenses/by-sa/3.0/nl/":        "Q18195572",
	"https://creativecommons.org/licenses/by-sa/2.5/nl/":        "Q18199175",
	"https://creativecommons.org/licenses/by/3.0/us/":           "Q18810143",
	"https://creativecommons.org/licenses/by-nd/3.0/":           "Q18810160",
	"https://creativecommons.org/licenses/by-nc/3.0/":           "Q18810331",
	"https://creativecommons.org/licenses/by/2.5/":              "Q18810333",
	"https://creativecommons.org/licenses/by-nd/2.5/":           "Q18810338",
	"https://creativecommons.org/licenses/by-sa/3.0/us/":        "Q18810341",
	"https://creativecommons.org/licenses/by-nc-nd/2.5/":        "Q19068204",
	"https://creativecommons.org/licenses/by-nc-sa/2.5/":        "Q19068212",
	"https://creativecommons.org/licenses/by-sa/2.0/":           "Q19068220",
	"https://creativecommons.org/licenses/by-nc/2.5/":           "Q19113746",
	"https://creativecommons.org/licenses/by-sa/2.5/":           "Q19113751",
	"https://creativecommons.org/licenses/by-nc-nd/3.0/":        "Q19125045",
	"https://creativecommons.org/licenses/by/2.0/":              "Q19125117",
	"https://creativecommons.org/licenses/by/4.0/":              "Q20007257",
	"https://creativecommons.org/licenses/by-nc-nd/4.0/":        "Q24082749",
	"https://creativecommons.org/licenses/by-sa/2.5/ca":         "Q24331618",
	"https://creativecommons.org/licenses/by/2.1/jp/":           "Q26116436",
	"https://creativecommons.org/licenses/by/3.0/igo/":          "Q26259495",
	"https://creativecommons.org/licenses/sampling+/1.0/":       "Q26913038",
	"https://creativecommons.org/licenses/by/2.5/se/":           "Q27940776",
	"https://creativecommons.org/licenses/by-nc-sa/2.0/":        "Q28050835",
	"https://creativecommons.org/licenses/by/1.0/":              "Q30942811",
	"https://creativecommons.org/licenses/by-nc/4.0/":           "Q34179348",
	"https://creativecommons.org/licenses/by-nd/2.0/":           "Q35254645",
	"https://creativecommons.org/licenses/by-nd/4.0/":           "Q36795408",
	"https://creativecommons.org/licenses/by-nc-nd/2.5/pt/":     "Q42172282",
	"https://creativecommons.org/licenses/by-nc-sa/4.0/":        "Q42553662",
	"https://creativecommons.org/licenses/by-sa/3.0/de/deed.de": "Q42716613",
	"https://creativecommons.org/licenses/by-nc/2.0":            "Q44128984",
	"https://creativecommons.org/licenses/by/2.0/kr":            "Q44282633",
	"https://creativecommons.org/licenses/by-sa/2.0/kr":         "Q44282641",
	"https://creativecommons.org/licenses/by-nc/1.0":            "Q44283370",
	"https://creativecommons.org/licenses/by-sa/1.0/":           "Q47001652",
	"https://creativecommons.org/licenses/by-nc-nd/1.0/":        "Q47008926",
	"https://creativecommons.org/licenses/by-nc-nd/2.0/":        "Q47008927",
	"https://creativecommons.org/licenses/by-nc-sa/1.0/":        "Q47008954",
	"https://creativecommons.org/licenses/by-nd/1.0/":           "Q47008966",
	"https://creativecommons.org/licenses/by/3.0/au/":           "Q52555753",
	"https://creativecommons.org/licenses/by/3.0/nl/":           "Q53859967",
	"https://creativecommons.org/licenses/by-sa/3.0/igo/":       "Q56292840",
	"https://creativecommons.org/licenses/by-nc-nd/2.0/uk/":     "Q56299316",
	"https://creativecommons.org/licenses/by-nc-sa/2.0/kr/":     "Q58041147",
}
