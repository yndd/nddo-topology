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

// Nddotopology struct
type Nddotopology struct {
	Topology []*NddotopologyTopology `json:"topology,omitempty"`
}

// NddotopologyTopology struct
type NddotopologyTopology struct {
	AdminState  *string                       `json:"admin-state,omitempty"`
	Defaults    *NddotopologyTopologyDefaults `json:"defaults,omitempty"`
	Description *string                       `json:"description,omitempty"`
	Kind        []*NddotopologyTopologyKind   `json:"kind,omitempty"`
	Link        []*NddotopologyTopologyLink   `json:"link,omitempty"`
	Name        *string                       `json:"name"`
	Node        []*NddotopologyTopologyNode   `json:"node,omitempty"`
	State       *NddotopologyTopologyState    `json:"state,omitempty"`
}

// NddotopologyTopologyDefaults struct
type NddotopologyTopologyDefaults struct {
	Tag []*NddotopologyTopologyDefaultsTag `json:"tag,omitempty"`
}

// NddotopologyTopologyDefaultsTag struct
type NddotopologyTopologyDefaultsTag struct {
	Key   *string `json:"key"`
	Value *string `json:"value,omitempty"`
}

// NddotopologyTopologyKind struct
type NddotopologyTopologyKind struct {
	Name *string                        `json:"name"`
	Tag  []*NddotopologyTopologyKindTag `json:"tag,omitempty"`
}

// NddotopologyTopologyKindTag struct
type NddotopologyTopologyKindTag struct {
	Key   *string `json:"key"`
	Value *string `json:"value,omitempty"`
}

// NddotopologyTopologyLink struct
type NddotopologyTopologyLink struct {
	AdminState  *string                              `json:"admin-state,omitempty"`
	Description *string                              `json:"description,omitempty"`
	Endpoints   []*NddotopologyTopologyLinkEndpoints `json:"endpoints,omitempty"`
	Name        *string                              `json:"name"`
	State       *NddotopologyTopologyLinkState       `json:"state,omitempty"`
	Tag         []*NddotopologyTopologyLinkTag       `json:"tag,omitempty"`
}

// NddotopologyTopologyLinkEndpoints struct
type NddotopologyTopologyLinkEndpoints struct {
	InterfaceName *string                                 `json:"interface-name"`
	NodeName      *string                                 `json:"node-name"`
	Tag           []*NddotopologyTopologyLinkEndpointsTag `json:"tag,omitempty"`
}

// NddotopologyTopologyLinkEndpointsTag struct
type NddotopologyTopologyLinkEndpointsTag struct {
	Key   *string `json:"key"`
	Value *string `json:"value,omitempty"`
}

// NddotopologyTopologyLinkState struct
type NddotopologyTopologyLinkState struct {
	LastUpdate *string                             `json:"last-update,omitempty"`
	Reason     *string                             `json:"reason,omitempty"`
	Status     *string                             `json:"status,omitempty"`
	Tag        []*NddotopologyTopologyLinkStateTag `json:"tag,omitempty"`
}

// NddotopologyTopologyLinkStateTag struct
type NddotopologyTopologyLinkStateTag struct {
	Key   *string `json:"key"`
	Value *string `json:"value,omitempty"`
}

// NddotopologyTopologyLinkTag struct
type NddotopologyTopologyLinkTag struct {
	Key   *string `json:"key"`
	Value *string `json:"value,omitempty"`
}

// NddotopologyTopologyNode struct
type NddotopologyTopologyNode struct {
	AdminState  *string                        `json:"admin-state,omitempty"`
	Description *string                        `json:"description,omitempty"`
	KindName    *string                        `json:"kind-name,omitempty"`
	Name        *string                        `json:"name"`
	State       *NddotopologyTopologyNodeState `json:"state,omitempty"`
	Tag         []*NddotopologyTopologyNodeTag `json:"tag,omitempty"`
}

// NddotopologyTopologyNodeState struct
type NddotopologyTopologyNodeState struct {
	Endpoint   []*NddotopologyTopologyNodeStateEndpoint `json:"endpoint,omitempty"`
	LastUpdate *string                                  `json:"last-update,omitempty"`
	Reason     *string                                  `json:"reason,omitempty"`
	Status     *string                                  `json:"status,omitempty"`
	Tag        []*NddotopologyTopologyNodeStateTag      `json:"tag,omitempty"`
}

// NddotopologyTopologyNodeStateEndpoint struct
type NddotopologyTopologyNodeStateEndpoint struct {
	Lag        *bool   `json:"lag,omitempty"`
	LagSubLink *bool   `json:"lag-sub-link,omitempty"`
	Name       *string `json:"name"`
}

// NddotopologyTopologyNodeStateTag struct
type NddotopologyTopologyNodeStateTag struct {
	Key   *string `json:"key"`
	Value *string `json:"value,omitempty"`
}

// NddotopologyTopologyNodeTag struct
type NddotopologyTopologyNodeTag struct {
	Key   *string `json:"key"`
	Value *string `json:"value,omitempty"`
}

// NddotopologyTopologyState struct
type NddotopologyTopologyState struct {
	LastUpdate *string                         `json:"last-update,omitempty"`
	Reason     *string                         `json:"reason,omitempty"`
	Status     *string                         `json:"status,omitempty"`
	Tag        []*NddotopologyTopologyStateTag `json:"tag,omitempty"`
}

// NddotopologyTopologyStateTag struct
type NddotopologyTopologyStateTag struct {
	Key   *string `json:"key"`
	Value *string `json:"value,omitempty"`
}

// Root is the root of the schema
type Root struct {
	TopoNddotopology *Nddotopology `json:"nddo-topology,omitempty"`
}
