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

	nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// TopologyFinalizer is the name of the finalizer added to
	// Topology to block delete operations until the physical node can be
	// deprovisioned.
	TopologyFinalizer string = "topology.topo.nddo.yndd.io"
)

// Topology struct
type Topology struct {
	// +kubebuilder:validation:Enum=`disable`;`enable`
	// +kubebuilder:default:="enable"
	AdminState *string           `json:"admin-state,omitempty"`
	Defaults   *TopologyDefaults `json:"defaults,omitempty"`
	// kubebuilder:validation:MinLength=1
	// kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[A-Za-z0-9 !@#$^&()|+=`~.,'/_:;?-]*"
	Description *string         `json:"description,omitempty"`
	Kind        []*TopologyKind `json:"kind,omitempty"`
	Name        *string         `json:"name"`
}

// TopologyDefaults struct
type TopologyDefaults struct {
	Tag []*TopologyDefaultsTag `json:"tag,omitempty"`
}

// TopologyDefaultsTag struct
type TopologyDefaultsTag struct {
	// kubebuilder:validation:MinLength=1
	// kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[A-Za-z0-9 !@#$^&()|+=`~.,'/_:;?-]*"
	Key *string `json:"key"`
	// kubebuilder:validation:MinLength=1
	// kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[A-Za-z0-9 !@#$^&()|+=`~.,'/_:;?-]*"
	Value *string `json:"value,omitempty"`
}

// TopologyKind struct
type TopologyKind struct {
	Name *string            `json:"name"`
	Tag  []*TopologyKindTag `json:"tag,omitempty"`
}

// TopologyKindTag struct
type TopologyKindTag struct {
	// kubebuilder:validation:MinLength=1
	// kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[A-Za-z0-9 !@#$^&()|+=`~.,'/_:;?-]*"
	Key *string `json:"key"`
	// kubebuilder:validation:MinLength=1
	// kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[A-Za-z0-9 !@#$^&()|+=`~.,'/_:;?-]*"
	Value *string `json:"value,omitempty"`
}

// TopologyParameters are the parameter fields of a Topology.
type TopologyParameters struct {
	TopoTopology *Topology `json:"topology,omitempty"`
}

// TopologyObservation are the observable fields of a Topology.
type TopologyObservation struct {
	//*Nddotopology `json:",inline"`
	TopoTopology *NddotopologyTopology `json:"topology,omitempty"`
}

// A TopologySpec defines the desired state of a Topology.
type TopologySpec struct {
	nddv1.ResourceSpec `json:",inline"`
	ForNetworkNode     TopologyParameters `json:"forNetworkNode"`
}

// A TopologyStatus represents the observed state of a Topology.
type TopologyStatus struct {
	nddv1.ResourceStatus `json:",inline"`
	AtNetworkNode        TopologyObservation `json:"atNetworkNode,omitempty"`
}

// +kubebuilder:object:root=true

// TopoTopology is the Schema for the Topology API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="TARGET",type="string",JSONPath=".status.conditions[?(@.kind=='TargetFound')].status"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.conditions[?(@.kind=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNC",type="string",JSONPath=".status.conditions[?(@.kind=='Synced')].status"
// +kubebuilder:printcolumn:name="LOCALLEAFREF",type="string",JSONPath=".status.conditions[?(@.kind=='InternalLeafrefValidationSuccess')].status"
// +kubebuilder:printcolumn:name="EXTLEAFREF",type="string",JSONPath=".status.conditions[?(@.kind=='ExternalLeafrefValidationSuccess')].status"
// +kubebuilder:printcolumn:name="PARENTDEP",type="string",JSONPath=".status.conditions[?(@.kind=='ParentValidationSuccess')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={ndd,topo}
type TopoTopology struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopologySpec   `json:"spec,omitempty"`
	Status TopologyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TopoTopologyList contains a list of Topologys
type TopoTopologyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TopoTopology `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TopoTopology{}, &TopoTopologyList{})
}

// Topology type metadata.
var (
	TopologyKindKind         = reflect.TypeOf(TopoTopology{}).Name()
	TopologyGroupKind        = schema.GroupKind{Group: Group, Kind: TopologyKindKind}.String()
	TopologyKindAPIVersion   = TopologyKindKind + "." + GroupVersion.String()
	TopologyGroupVersionKind = GroupVersion.WithKind(TopologyKindKind)
)
