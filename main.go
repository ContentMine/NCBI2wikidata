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
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello, world")

	f, err := os.Open("fetch1.xml")
	if err != nil {
		panic(err)
	}

	a := PubmedArticleSet{}
	err = xml.NewDecoder(f).Decode(&a)
	if err != nil {
		panic(err)
	}


	for _, article := range a.Articles {
	    fmt.Printf("\n%s\n----\n", article.MedlineCitation.Article[0].ArticleTitle)
	    for _, md := range article.MedlineCitation.MeshHeadingList.MeshHeadings {
	        flag := ""
	        if md.DescriptorName.MajorTopicYN == "Y" {
	            flag = "*"
	        }
	        fmt.Printf("\t%s %s\n", md.DescriptorName.Name, flag)
	        for _, q := range md.QualifierNames {
                flag := ""
                if q.MajorTopicYN == "Y" {
                    flag = "*"
                }
	            fmt.Printf("\t\t%s %s\n", q.Name, flag)
	        }
	    }
	}

}
