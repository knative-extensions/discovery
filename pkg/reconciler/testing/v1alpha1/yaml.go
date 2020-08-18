/*
Copyright 2020 The Knative Authors.

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

package v1alpha1

import (
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func ParseYAML(yamls io.Reader) ([]unstructured.Unstructured, error) {
	decoder := yaml.NewYAMLToJSONDecoder(yamls)
	var objs []unstructured.Unstructured
	var err error
	for {
		out := unstructured.Unstructured{}
		err = decoder.Decode(&out)
		if err != nil {
			break
		}
		objs = append(objs, out)
	}
	if err != io.EOF {
		return nil, err
	}
	return objs, nil
}
