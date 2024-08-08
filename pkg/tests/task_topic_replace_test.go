/*
 * Copyright 2024 InfAI (CC SES)
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

package tests

import (
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/controller"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/controller/etree"
	"github.com/SENERGY-Platform/mgw-process-sync-client/pkg/tests/resources"
	"strings"
	"testing"
)

func TestTaskTopicReplace(t *testing.T) {
	replacedByStrRepl := strings.ReplaceAll(resources.IncidentBpmn, "optimistic", "pessimistic")
	replaced, err := controller.ReplaceTaskTopics(resources.IncidentBpmn, map[string]string{"optimistic": "pessimistic"})
	if err != nil {
		t.Error(err)
		return
	}
	if !bpmncompare(t, replaced, replacedByStrRepl) {
		t.Errorf("replaced != replacedByStrRepl\n%#v\n%#v\n", replaced, replacedByStrRepl)
	}
	reverse, err := controller.ReplaceTaskTopics(replaced, map[string]string{"pessimistic": "optimistic"})
	if !bpmncompare(t, reverse, resources.IncidentBpmn) {
		t.Errorf("reverse != resources.IncidentBpmn\n%#v\n%#v\n", reverse, resources.IncidentBpmn)
	}
}

func bpmncompare(t *testing.T, a string, b string) bool {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	doca := etree.NewDocument()
	err := doca.ReadFromString(a)
	if err != nil {
		t.Error(err)
		return false
	}
	docb := etree.NewDocument()
	err = docb.ReadFromString(b)
	if err != nil {
		t.Error(err)
		return false
	}

	norma, err := doca.WriteToString()
	if err != nil {
		t.Error(err)
		return false
	}
	normb, err := docb.WriteToString()
	if err != nil {
		t.Error(err)
		return false
	}
	return norma == normb
}
