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
)

type Source struct {
	ID    string
	Value string
}

type AddStatement struct {
	ItemID     string
	PropertyID string
	Value      string
	SourceList []Source
}

func (a *AddStatement) String() string {
	statement := fmt.Sprintf("%s\t%s\t%s", a.ItemID, a.PropertyID, a.Value)
	for _, source := range a.SourceList {
		statement = fmt.Sprintf("%s\t%s\t%s", statement, source.ID, source.Value)
	}
	return statement + "\n"
}

func AddItemPropertyToItem(target_id string, property_id string, value_id string) *AddStatement {
	return &AddStatement{
		ItemID:     target_id,
		PropertyID: property_id,
		Value:      value_id,
		SourceList: make([]Source, 0),
	}
}

func AddStringPropertyToItem(target_id string, property_id string, value string) *AddStatement {
	return &AddStatement{
		ItemID:     target_id,
		PropertyID: property_id,
		Value:      value,
		SourceList: make([]Source, 0),
	}
}

func (a *AddStatement) AddSource(source_id, value string) {
	a.SourceList = append(a.SourceList, Source{ID: source_id, Value: value})
}
