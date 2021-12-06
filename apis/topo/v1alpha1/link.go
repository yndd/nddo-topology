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

func (x *TopologyLinkParameters) GetTopologyName() string {
	if reflect.ValueOf(x.TopologyName).IsZero() {
		return ""
	}
	return *x.TopologyName
}

func (x *NddotopologyTopologyLink) GetName() string {
	if reflect.ValueOf(x.Name).IsZero() {
		return ""
	}
	return *x.Name
}

func (x *NddotopologyTopologyLink) GetAdminState() string {
	if reflect.ValueOf(x.AdminState).IsZero() {
		return ""
	}
	return *x.AdminState
}

func (x *NddotopologyTopologyLink) GetDescription() string {
	if reflect.ValueOf(x.Description).IsZero() {
		return ""
	}
	return *x.Description
}

func (x *NddotopologyTopologyLink) GetStatus() string {
	if reflect.ValueOf(x.State.Status).IsZero() {
		return ""
	}
	return *x.State.Status
}

func (x *NddotopologyTopologyLink) GetReason() string {
	if reflect.ValueOf(x.State.Reason).IsZero() {
		return ""
	}
	return *x.State.Reason
}

func (x *NddotopologyTopologyLink) GetLag() bool {
	if reflect.ValueOf(x.Tag).IsZero() {
		return false
	}
	for _, tag := range x.Tag {
		if *tag.Key == Lag {
			// found
			return *tag.Value == "true"
		}
	}
	// not found
	return false
}

func (n *NddotopologyTopologyLink) SetDescription(s string) {
	n.Description = &s
}

func (n *NddotopologyTopologyLink) SetStatus(s string) {
	n.State.Status = &s
}

func (n *NddotopologyTopologyLink) SetReason(s string) {
	n.State.Reason = &s
}

func (n *NddotopologyTopologyLink) SetLastUpdate(s string) {
	n.State.LastUpdate = &s
}
