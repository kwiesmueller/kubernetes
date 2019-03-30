/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handlers

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestManagerOrUserAgent(t *testing.T) {
	tests := []struct {
		manager   string
		userAgent string
		expected  string
	}{
		{
			manager:   "",
			userAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36",
			expected:  "Mozilla",
		},
		{
			manager:   "",
			userAgent: "fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff/Something",
			expected:  "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		},
		{
			manager:   "",
			userAgent: "🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔",
			expected:  "🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔🍔",
		},
		{
			manager:   "",
			userAgent: "userAgent",
			expected:  "userAgent",
		},
		{
			manager:   "manager",
			userAgent: "userAgent",
			expected:  "manager",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v-%v", test.manager, test.userAgent), func(t *testing.T) {
			got := fieldManagerOrUserAgent(test.manager, prefixFromUserAgent(test.userAgent))
			if got != test.expected {
				t.Errorf("Wanted %#v, got %#v", test.expected, got)
			}
		})
	}
}

func TestValidFieldManager(t *testing.T) {
	tests := []struct {
		manager  runtime.Object
		field    string
		expected string
	}{
		{
			manager:  &unstructured.Unstructured{},
			field:    "",
			expected: "",
		},
		{
			manager:  &unstructured.Unstructured{},
			field:    "field",
			expected: "field",
		},
		{
			manager: &unstructured.Unstructured{Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"fieldManager": "manager",
				},
			}},
			field:    "",
			expected: "manager",
		},
		{
			manager: &unstructured.Unstructured{Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"fieldManager": "manager",
				},
			}},
			field:    "manager",
			expected: "manager",
		},
		{
			manager: &unstructured.Unstructured{Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"fieldManager": "manager",
				},
			}},
			field:    "field",
			expected: "",
		},
		{
			manager: &unstructured.Unstructured{Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"fieldManager": strings.Repeat("f", validation.FieldManagerMaxLength+1),
				},
			}},
			field:    "",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v-%v", test.manager, test.field), func(t *testing.T) {
			got, _ := validFieldManager(test.manager, test.field)
			if got != test.expected {
				t.Errorf("Wanted %#v, got %#v", test.expected, got)
			}
		})
	}
}

func TestManagerValidation(t *testing.T) {
	tests := []struct {
		manager     runtime.Object
		field       string
		expectedErr bool
	}{
		{
			manager: &unstructured.Unstructured{Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"fieldManager": "manager",
				},
			}},
			field:       "field",
			expectedErr: true,
		},
		{
			manager: &unstructured.Unstructured{Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"fieldManager": strings.Repeat("f", validation.FieldManagerMaxLength+1),
				},
			}},
			field:       "",
			expectedErr: true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v-%v-%v", test.manager, test.field), func(t *testing.T) {
			_, err := validFieldManager(test.manager, test.field)
			if (err == nil) == test.expectedErr {
				t.Errorf("Did not get expected err: %#v", err)
			}
		})
	}
}
