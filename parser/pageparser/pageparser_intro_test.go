// Copyright 2018 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pageparser

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type lexerTest struct {
	name  string
	input string
	items []Item
}

func nti(tp itemType, val string) Item {
	return Item{tp, 0, []byte(val)}
}

var (
	tstJSON                = `{ "a": { "b": "\"Hugo\"}" } }`
	tstHTMLLead            = nti(tHTMLLead, "  <")
	tstFrontMatterTOML     = nti(tFrontMatterTOML, "foo = \"bar\"\n")
	tstFrontMatterYAML     = nti(tFrontMatterYAML, "foo: \"bar\"\n")
	tstFrontMatterYAMLCRLF = nti(tFrontMatterYAML, "foo: \"bar\"\r\n")
	tstFrontMatterJSON     = nti(tFrontMatterJSON, tstJSON+"\r\n")
	tstSomeText            = nti(tText, "\nSome text.\n")
	tstSummaryDivider      = nti(tSummaryDivider, "<!--more-->")
	tstSummaryDividerOrg   = nti(tSummaryDividerOrg, "# more")

	tstORG = `
#+TITLE: T1
#+AUTHOR: A1
#+DESCRIPTION: D1
`
	tstFrontMatterORG = nti(tFrontMatterORG, tstORG)
)

var crLfReplacer = strings.NewReplacer("\r", "#", "\n", "$")

// TODO(bep) a way to toggle ORG mode vs the rest.
var frontMatterTests = []lexerTest{
	{"empty", "", []Item{tstEOF}},
	{"HTML Document", `  <html>  `, []Item{tstHTMLLead, nti(tText, "html>  "), tstEOF}},
	{"YAML front matter", "---\nfoo: \"bar\"\n---\n\nSome text.\n", []Item{tstFrontMatterYAML, tstSomeText, tstEOF}},
	// Note that we keep all bytes as they are, but we need to handle CRLF
	{"YAML front matter CRLF", "---\r\nfoo: \"bar\"\r\n---\n\nSome text.\n", []Item{tstFrontMatterYAMLCRLF, tstSomeText, tstEOF}},
	{"TOML front matter", "+++\nfoo = \"bar\"\n+++\n\nSome text.\n", []Item{tstFrontMatterTOML, tstSomeText, tstEOF}},
	{"JSON front matter", tstJSON + "\r\n\nSome text.\n", []Item{tstFrontMatterJSON, tstSomeText, tstEOF}},
	{"ORG front matter", tstORG + "\nSome text.\n", []Item{tstFrontMatterORG, tstSomeText, tstEOF}},
	{"Summary divider ORG", tstORG + "\nSome text.\n# more\nSome text.\n", []Item{tstFrontMatterORG, tstSomeText, tstSummaryDividerOrg, tstSomeText, tstEOF}},
	{"Summary divider", "+++\nfoo = \"bar\"\n+++\n\nSome text.\n<!--more-->\nSome text.\n", []Item{tstFrontMatterTOML, tstSomeText, tstSummaryDivider, tstSomeText, tstEOF}},
}

func TestFrontMatter(t *testing.T) {
	t.Parallel()
	for i, test := range frontMatterTests {
		items := collect([]byte(test.input), false, lexIntroSection)
		if !equal(items, test.items) {
			got := crLfReplacer.Replace(fmt.Sprint(items))
			expected := crLfReplacer.Replace(fmt.Sprint(test.items))
			t.Errorf("[%d] %s: got\n\t%v\nexpected\n\t%v", i, test.name, got, expected)
		}
	}
}

func collect(input []byte, skipFrontMatter bool, stateStart stateFunc) (items []Item) {
	l := newPageLexer(input, 0, stateStart)
	l.run()

	for {
		item := l.nextItem()
		items = append(items, item)
		if item.typ == tEOF || item.typ == tError {
			break
		}
	}
	return
}

// no positional checking, for now ...
func equal(i1, i2 []Item) bool {
	if len(i1) != len(i2) {
		return false
	}
	for k := range i1 {
		if i1[k].typ != i2[k].typ {
			return false
		}
		if !reflect.DeepEqual(i1[k].Val, i2[k].Val) {
			return false
		}
	}
	return true
}
