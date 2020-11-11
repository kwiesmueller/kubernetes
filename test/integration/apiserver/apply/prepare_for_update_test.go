package apiserver

import (
	"context"
	"reflect"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	genericfeatures "k8s.io/apiserver/pkg/features"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
)

func TestPrepareForUpdate(t *testing.T) {
	defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, genericfeatures.ServerSideApply, true)()

	_, client, closeFn := setup(t)
	defer closeFn()

	testCases := []struct {
		resource string
		name     string
		body     string
	}{
		{
			resource: "pods",
			name:     "test-pod-with-status",
			body: `{
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"name": "test-pod-with-status"
				},
				"spec": {
					"containers": [{
						"name":  "test-container",
						"image": "test-image"
					}]
				},
				"status": {
					"phase": "testing"
				}
			}`,
		},
	}

	for _, tc := range testCases {
		_, err := client.CoreV1().RESTClient().Patch(types.ApplyPatchType).
			Namespace("default").
			Resource(tc.resource).
			Name(tc.name).
			Param("fieldManager", "apply_test").
			Body([]byte(tc.body)).
			Do(context.TODO()).
			Get()
		if err != nil {
			t.Fatalf("Failed to create object using Apply patch: %v", err)
		}

		pod, err := client.CoreV1().Pods("default").Get(context.TODO(), tc.name, v1.GetOptions{})
		if err != nil {
			t.Fatalf("Failed to retrieve object: %v", err)
		}

		if string(pod.Status.Phase) == "testing" {
			t.Fatalf("Pod should not have .status.phase 'testing'")
		}

		actualFieldsV1 := pod.ObjectMeta.ManagedFields[0].FieldsV1.Raw
		expectedFieldsV1 := []byte(`{"f:spec":{"f:containers":{"k:{\"name\":\"test-container\"}":{".":{},"f:image":{},"f:name":{}}}}}`)

		if !reflect.DeepEqual(actualFieldsV1, expectedFieldsV1) {
			t.Fatalf("expected managedFields to be:\n%s\ngot:\n%s", string(expectedFieldsV1), string(actualFieldsV1))
		}
	}
}
