/*
Copyright 2021 NDD.

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

import "reflect"

func (x *NddotopologyTopologyLinkEndpoints) GetNodeName() string {
	if reflect.ValueOf(x.NodeName).IsZero() {
		return ""
	}
	return *x.NodeName
}

func (x *NddotopologyTopologyLinkEndpoints) GetInterfaceName() string {
	if reflect.ValueOf(x.InterfaceName).IsZero() {
		return ""
	}
	return *x.InterfaceName
}

func (x *NddotopologyTopologyLinkEndpoints) GetLagName() string {
	if reflect.ValueOf(x.Tag).IsZero() {
		return ""
	}
	for _, tag := range x.Tag {
		if *tag.Key == LagName {
			// found
			return *tag.Value
		}
	}
	// not found
	return ""
}
