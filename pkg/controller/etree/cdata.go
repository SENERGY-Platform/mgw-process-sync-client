/*
 * Copyright 2021 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package etree

import (
	"bufio"
)

type CData struct {
	parent *Element
	Value  string
}

func (this *CData) Parent() *Element {
	return this.parent
}

func (this *CData) setParent(parent *Element) {
	this.parent = parent
}

func (this *CData) writeTo(w *bufio.Writer, s *WriteSettings) {
	w.WriteString("<![CDATA[" + this.Value + "]]>")
}

func (this *CData) dup(parent *Element) Token {
	return &CData{
		parent: parent,
		Value:  this.Value,
	}
}

func (e *Element) WriteCData(text string) {
	e.Child = []Token{&CData{parent: e, Value: text}}
}
