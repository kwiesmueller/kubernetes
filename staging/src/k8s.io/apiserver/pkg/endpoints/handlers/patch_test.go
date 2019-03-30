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
	"errors"
	"net/http"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/endpoints/handlers/fieldmanager"
)

type fakeObjectConvertor struct{}

func (c *fakeObjectConvertor) Convert(in, out, context interface{}) error {
	out = in
	return nil
}

func (c *fakeObjectConvertor) ConvertToVersion(in runtime.Object, _ runtime.GroupVersioner) (runtime.Object, error) {
	return in, nil
}

func (c *fakeObjectConvertor) ConvertFieldLabel(_ schema.GroupVersionKind, _, _ string) (string, string, error) {
	return "", "", errors.New("not implemented")
}

type fakeObjectDefaulter struct{}

func (d *fakeObjectDefaulter) Default(in runtime.Object) {}

func TestApplyPatchVersionCheck(t *testing.T) {
	gvk := schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Pod",
	}

	p := &applyPatcher{
		options: &metav1.PatchOptions{
			FieldManager: "test",
		},
		kind: gvk,
		fieldManager: fieldmanager.NewCRDFieldManager(
			&fakeObjectConvertor{},
			&fakeObjectDefaulter{},
			gvk.GroupVersion(),
			gvk.GroupVersion(),
		),
	}

	obj := &corev1.Pod{}
	p.patch = []byte(`{
		"apiVersion": "apps/v1",
		"kind": "Pod",
	}`)

	// patch has 'apiVersion: apps/v1' and live version is apps/v1 -> no errors
	_, err := p.applyPatchToCurrentObject(obj)
	if err != nil {
		t.Fatalf("failed to apply object: %v", err)
	}

	p.patch = []byte(`{
		"apiVersion": "apps/v2",
		"kind": "Pod",
	}`)

	// patch has 'apiVersion: apps/v2' but live version is apps/v1 -> error
	_, err = p.applyPatchToCurrentObject(obj)
	if err == nil {
		t.Fatalf("expected an error from mismatched patch and live versions")
	}
	switch typ := err.(type) {
	default:
		t.Fatalf("expected error to be of type %T was %T", apierrors.StatusError{}, typ)
	case apierrors.APIStatus:
		if typ.Status().Code != http.StatusBadRequest {
			t.Fatalf("expected status code to be %d but was %d",
				http.StatusBadRequest, typ.Status().Code)
		}
	}
}
