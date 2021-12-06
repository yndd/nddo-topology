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

import (
	"reflect"

	"github.com/yndd/ndd-runtime/pkg/utils"
)

func (x *NddotopologyTopologyNode) GetName() string {
	if reflect.ValueOf(x.Name).IsZero() {
		return ""
	}
	return *x.Name
}

func (x *NddotopologyTopologyNode) GetAdminState() string {
	if reflect.ValueOf(x.AdminState).IsZero() {
		return ""
	}
	return *x.AdminState
}

func (x *NddotopologyTopologyNode) GetDescription() string {
	if reflect.ValueOf(x.Description).IsZero() {
		return ""
	}
	return *x.Description
}

func (x *NddotopologyTopologyNode) GetKindName() string {
	if reflect.ValueOf(x.KindName).IsZero() {
		return ""
	}
	return *x.KindName
}

func (x *NddotopologyTopologyNode) GetStatus() string {
	if reflect.ValueOf(x.State.Status).IsZero() {
		return ""
	}
	return *x.State.Status
}

func (x *NddotopologyTopologyNode) GetReason() string {
	if reflect.ValueOf(x.State.Reason).IsZero() {
		return ""
	}
	return *x.State.Reason
}

func (x *NddotopologyTopologyNode) GetPlatform() string {
	if reflect.ValueOf(x.Tag).IsZero() {
		return ""
	}
	for _, tag := range x.Tag {
		if *tag.Key == Platform {
			// found
			return *tag.Value
		}
	}
	// not found
	return ""
}

func (x *NddotopologyTopologyNode) GetEndpoints() []*NddotopologyTopologyNodeStateEndpoint {
	if x.State != nil && x.State.Endpoint != nil {
		return x.State.Endpoint
	}
	// not found
	return nil
}

func (x *NddotopologyTopologyNode) GetEndpoint(n string) *NddotopologyTopologyNodeStateEndpoint {
	if x.State != nil && x.State.Endpoint != nil {
		for _, ep := range x.State.Endpoint {
			if ep.GetName() == n {
				return ep
			}
		}
	}
	// not found
	return nil
}

func (x *NddotopologyTopologyNode) SetEndpoint(e *NddotopologyTopologyNodeStateEndpoint) {
	if x.State != nil && x.State.Endpoint != nil {
		found := false
		for _, ep := range x.State.Endpoint {
			if ep.GetName() == e.GetName() {
				// update the endpoint
				ep = e
				found = true
				break
			}
		}
		if !found {
			x.State.Endpoint = append(x.State.Endpoint, e)
		}
	}
	// we should never come here
}

func (n *NddotopologyTopologyNode) SetDescription(s string) {
	n.Description = &s
}

func (n *NddotopologyTopologyNode) SetStatus(s string) {
	n.State.Status = &s
}

func (n *NddotopologyTopologyNode) SetReason(s string) {
	n.State.Reason = &s
}

func (n *NddotopologyTopologyNode) SetLastUpdate(s string) {
	n.State.LastUpdate = &s
}

func (n *NddotopologyTopologyNode) SetPlatform(s string) {
	if n.Tag == nil {
		n.Tag = make([]*NddotopologyTopologyNodeTag, 0)
	}
	t := &NddotopologyTopologyNodeTag{
		Key:   utils.StringPtr(Platform),
		Value: &s,
	}
	n.Tag = append(n.Tag, t)
}
