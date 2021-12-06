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
package applogic

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/karimra/gnmic/types"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/pkg/errors"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-yang/pkg/cache"
	"github.com/yndd/ndd-yang/pkg/dispatcher"
	"github.com/yndd/ndd-yang/pkg/yentry"
	"github.com/yndd/ndd-yang/pkg/yparser"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type root struct {
	dispatcher.Resource
	data       interface{}
	topologies map[string]dispatcher.Handler
	target     *targetInfo
}

func (r *root) WithLogging(log logging.Logger) {
	r.Log = log
}

func (r *root) WithStateCache(c *cache.Cache) {
	r.StateCache = c
}

func (r *root) WithConfigCache(c *cache.Cache) {
	r.ConfigCache = c
}

func (r *root) WithTargetCache(c *cache.Cache) {
	r.TargetCache = c
}

func (r *root) WithPrefix(p *gnmi.Path) {
	r.Prefix = p
}

func (r *root) WithPathElem(pe []*gnmi.PathElem) {
	r.PathElem = pe[0]
}

func (r *root) WithRootSchema(rs *yentry.Entry) {
	r.RootSchema = rs
}

func (r *root) WithK8sClient(c client.Client) {
	r.Client = c
}

func NewRoot(opts ...dispatcher.Option) dispatcher.Handler {
	r := &root{
		topologies: make(map[string]dispatcher.Handler),
	}

	for _, opt := range opts {
		opt(r)
	}

	// initialize target before any config event is received
	r.target = NewTarget(
		WithLogger(r.Log),
		WithK8sClient(r.Client),
	)

	return r
}

func (r *root) HandleConfigEvent(o dispatcher.Operation, prefix *gnmi.Path, pe []*gnmi.PathElem, d interface{}) (dispatcher.Handler, error) {
	log := r.Log.WithValues("Operation", o, "Path Elem", pe)

	log.Debug("HandleConfigEvent root", "Data", d)

	children := map[string]dispatcher.HandleConfigEventFunc{
		"topology": topologyCreate,
	}

	// check path Element Name
	pathElemName := pe[0].GetName()
	if _, ok := children[pathElemName]; !ok {
		return nil, errors.Wrap(errors.New("unexpected pathElem"), fmt.Sprintf("root handle: %s", pathElemName))
	}

	if len(pe) == 1 {
		//log.Debug("root handle pathelem =1")
		// handle local
		switch o {
		case dispatcher.OperationUpdate:
			i, err := r.CreateChild(children, pathElemName, prefix, pe, d)
			if err != nil {
				return nil, err
			}
			//r.Log.Debug("topology update", "data", d)
			if d != nil {
				if err := i.UpdateConfig(d); err != nil {
					return nil, err
				}
				if err := i.UpdateStateCache(); err != nil {
					return nil, err
				}
			}
			return i, nil
		case dispatcher.OperationDelete:
			if err := r.DeleteChild(pathElemName, pe); err != nil {
				return nil, err
			}
			return nil, nil
		}
	} else {
		//log.Debug("root Handle pathelem >1")
		i, err := r.CreateChild(children, pathElemName, prefix, pe[:1], nil)
		if err != nil {
			return nil, err
		}
		return i.HandleConfigEvent(o, prefix, pe[1:], d)
	}
	return nil, nil
}

func (r *root) CreateChild(children map[string]dispatcher.HandleConfigEventFunc, pathElemName string, prefix *gnmi.Path, pe []*gnmi.PathElem, d interface{}) (dispatcher.Handler, error) {

	switch pathElemName {
	case "topology":
		if i, ok := r.topologies[topologyGetKey(pe)]; !ok {
			i = children[pathElemName](r.Log, r.ConfigCache, r.StateCache, r.TargetCache, r.Client, prefix, pe, d)
			i.SetRootSchema(r.RootSchema)
			if err := i.SetParent(r); err != nil {
				return nil, err
			}
			r.topologies[topologyGetKey(pe)] = i
			return i, nil
		} else {
			return i, nil
		}
	}

	return nil, errors.New("CreateChild unexpected pathElemName in root")
}

func (r *root) DeleteChild(pathElemName string, pe []*gnmi.PathElem) error {

	switch pathElemName {
	case "topology":
		if i, ok := r.topologies[topologyGetKey(pe)]; ok {
			if err := i.DeleteStateCache(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *root) SetParent(parent interface{}) error {
	// no SetParent required for root
	return nil
}

func (r *root) SetRootSchema(rs *yentry.Entry) {
	r.RootSchema = rs
}

func (r *root) GetChildren() map[string]string {
	x := make(map[string]string)
	for k := range r.topologies {
		x[k] = "topology"
	}
	return x
}

func (r *root) UpdateConfig(d interface{}) error {
	// no updates required for root
	return nil
}

func (r *root) GetPathElem(p []*gnmi.PathElem, do_recursive bool) ([]*gnmi.PathElem, error) {
	return nil, nil
}

func (r *root) UpdateStateCache() error {
	pe, err := r.GetPathElem(nil, true)
	if err != nil {
		return err
	}
	b, err := json.Marshal(r.data)
	if err != nil {
		return err
	}
	var x interface{}
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}
	//r.Log.Debug("Debug updateState", "data", x)
	u, err := yparser.GetGranularUpdatesFromJSON(&gnmi.Path{Elem: pe}, x, r.RootSchema)
	n := &gnmi.Notification{
		Timestamp: time.Now().UnixNano(),
		Prefix:    r.Prefix,
		Update:    u,
	}
	if err != nil {
		return err
	}
	if u != nil {
		if err := r.StateCache.GnmiUpdate(r.Prefix.Target, n); err != nil {
			if strings.Contains(fmt.Sprintf("%v", err), "stale") {
				return nil
			}
			return err
		}
	}
	return nil
}

func (r *root) DeleteStateCache() error {
	pe, err := r.GetPathElem(nil, true)
	if err != nil {
		return err
	}
	n := &gnmi.Notification{
		Timestamp: time.Now().UnixNano(),
		Prefix:    r.Prefix,
		Delete: []*gnmi.Path{
			{Elem: pe},
		},
	}
	if err := r.StateCache.GnmiUpdate(r.Prefix.Target, n); err != nil {
		return err
	}
	return nil
}

func (r *root) GetData(resource string, key map[string]string) (interface{}, error) {
	return nil, nil
}

func (r *root) SetData(resource string, key map[string]string, d interface{}) error {
	return nil
}

func (r *root) GetTarget() *targetInfo {
	return r.target
}

func (r *root) GetTargets() []*types.TargetConfig {
	return r.target.GetTargets()
}
