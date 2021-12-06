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

func (x *NddotopologyTopology) GetName() string {
	if reflect.ValueOf(x.Name).IsZero() {
		return ""
	}
	return *x.Name
}

func (x *NddotopologyTopology) GetAdminState() string {
	if reflect.ValueOf(x.AdminState).IsZero() {
		return ""
	}
	return *x.AdminState
}

func (x *NddotopologyTopology) GetDescription() string {
	if reflect.ValueOf(x.Description).IsZero() {
		return ""
	}
	return *x.Description
}

func (x *NddotopologyTopology) GetStatus() string {
	if x.State == nil {
		return ""
	}
	if reflect.ValueOf(x.State.Status).IsZero() {
		return ""
	}
	return *x.State.Status
}

func (x *NddotopologyTopology) GetReason() string {
	if x.State == nil {
		return ""
	}
	if reflect.ValueOf(x.State.Reason).IsZero() {
		return ""
	}
	return *x.State.Reason
}

func (t *NddotopologyTopology) GetDefaults() *NddotopologyTopologyDefaults {
	return t.Defaults
}

func (t *NddotopologyTopology) GetKinds() []*NddotopologyTopologyKind {
	return t.Kind
}

func (n *NddotopologyTopology) SetDescription(s string) {
	n.Description = &s
}

func (n *NddotopologyTopology) SetStatus(s string) {
	n.State.Status = &s
}

func (n *NddotopologyTopology) SetReason(s string) {
	n.State.Reason = &s
}

func (n *NddotopologyTopology) SetLastUpdate(s string) {
	n.State.LastUpdate = &s
}
