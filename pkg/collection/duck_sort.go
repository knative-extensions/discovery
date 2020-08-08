/*
Copyright 2020 The Knative Authors

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

package collection

import (
	"fmt"

	"knative.dev/discovery/pkg/apis/discovery/v1alpha1"
)

// ByResourceMeta implements sort.Interface for []v1alpha1.ResourceMeta based on
// the group and resource fields.
type ByResourceMeta []v1alpha1.ResourceMeta

func (a ByResourceMeta) Len() int      { return len(a) }
func (a ByResourceMeta) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByResourceMeta) Less(i, j int) bool {
	// TODO: this needs to sort like this, but also group by duck version.
	keyI := fmt.Sprintf("%s-%s", a[i].APIVersion, a[i].Kind)
	keyJ := fmt.Sprintf("%s-%s", a[j].APIVersion, a[j].Kind)
	return keyI < keyJ
}
