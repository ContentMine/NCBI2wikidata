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
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello, world")

	request := ESearchRequest{
		Term:   "food",
		DB:     "pubmed",
		APIKey: os.Getenv("NCBI_API_KEY"),
	}

	res, err := request.Do()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response %v\n", res)
}
