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

package internal

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

// RemoveObjectFieldManager removes the manager field from the object
// so we don't persist it.
func RemoveObjectFieldManager(obj runtime.Object) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		panic(fmt.Sprintf("couldn't get accessor: %v", err))
	}
	accessor.SetFieldManager("")
}
