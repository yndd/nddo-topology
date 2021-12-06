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
	topov1alpha1 "github.com/yndd/nddo-topology/apis/topo/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type node struct {
	dispatcher.Resource
	data   *topov1alpha1.NddotopologyTopologyNode
	parent *topology
	//activeTargets []*types.TargetConfig // active targets
	//target        *targetInfo           // observer interface
	//ctx           context.Context
}

func (r *node) WithLogging(log logging.Logger) {
	r.Log = log
}

func (r *node) WithStateCache(c *cache.Cache) {
	r.StateCache = c
}

func (r *node) WithTargetCache(c *cache.Cache) {
	r.TargetCache = c
}

func (r *node) WithConfigCache(c *cache.Cache) {
	r.ConfigCache = c
}

func (r *node) WithPrefix(p *gnmi.Path) {
	r.Prefix = p
}

func (r *node) WithPathElem(pe []*gnmi.PathElem) {
	r.PathElem = pe[0]
}

func (r *node) WithRootSchema(rs *yentry.Entry) {
	r.RootSchema = rs
}

func (r *node) WithK8sClient(c client.Client) {
	r.Client = c
}

func NewNode(n string, opts ...dispatcher.Option) dispatcher.Handler {
	x := &node{}

	x.Key = n

	for _, opt := range opts {
		opt(x)
	}

	return x
}

func nodeGetKey(p []*gnmi.PathElem) string {
	return p[0].GetKey()["name"]
}

func nodeCreate(log logging.Logger, cc, sc, tc *cache.Cache, c client.Client, prefix *gnmi.Path, p []*gnmi.PathElem, d interface{}) dispatcher.Handler {
	nodeName := nodeGetKey(p)
	return NewNode(nodeName,
		dispatcher.WithPrefix(prefix),
		dispatcher.WithPathElem(p),
		dispatcher.WithLogging(log),
		dispatcher.WithStateCache(sc),
		dispatcher.WithTargetCache(tc),
		dispatcher.WithConfigCache(cc),
		dispatcher.WithK8sClient(c))
}

func (r *node) HandleConfigEvent(o dispatcher.Operation, prefix *gnmi.Path, pe []*gnmi.PathElem, d interface{}) (dispatcher.Handler, error) {
	log := r.Log.WithValues("Operation", o, "Path Elem", pe)

	log.Debug("HandleConfigEvent node")

	if len(pe) == 1 {
		return nil, errors.New("the handle should have been terminated in the parent")
	} else {
		return nil, errors.New("there is no children in the node resource ")
	}
}

func (r *node) CreateChild(children map[string]dispatcher.HandleConfigEventFunc, pathElemName string, prefix *gnmi.Path, pe []*gnmi.PathElem, d interface{}) (dispatcher.Handler, error) {
	return nil, errors.New("there is no children in the node resource ")
}

func (r *node) DeleteChild(pathElemName string, pe []*gnmi.PathElem) error {
	return errors.New("there is no children in the node resource ")
}

func (r *node) SetParent(parent interface{}) error {
	p, ok := parent.(*topology)
	if !ok {
		return errors.New("wrong parent object")
	}
	r.parent = p

	r.parent.RegisterTopologyObserver(r)
	return nil
}

func (r *node) SetRootSchema(rs *yentry.Entry) {
	r.RootSchema = rs
}

func (r *node) GetChildren() map[string]string {
	x := make(map[string]string)
	return x
}

func (r *node) UpdateConfig(d interface{}) error {
	r.Copy(d)
	// from the moment we get dat we are interested in target updates
	//r.GetTarget().RegisterTargetObserver(r)
	// add channel here
	if err := r.handleConfigUpdate(); err != nil {
		return err
	}

	return nil
}

func (r *node) GetPathElem(p []*gnmi.PathElem, do_recursive bool) ([]*gnmi.PathElem, error) {
	//r.Log.Debug("GetPathElem", "PathElem node", r.PathElem)
	if r.parent != nil {
		p, err := r.parent.GetPathElem(p, true)
		if err != nil {
			return nil, err
		}
		p = append(p, r.PathElem)
		return p, nil
	}
	return nil, nil
}

func (r *node) Copy(d interface{}) error {
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	x := topov1alpha1.NddotopologyTopologyNode{}
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}
	r.data = (&x).DeepCopy()
	//r.Log.Debug("Copy", "Data", r.data)
	return nil
}

func (r *node) UpdateStateCache() error {
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
	//log.Debug("Debug updateState", "refPaths", refPaths)
	//r.Log.Debug("Debug updateState", "data", x)
	u, err := yparser.GetGranularUpdatesFromJSON(&gnmi.Path{Elem: pe}, x, r.RootSchema)
	n := &gnmi.Notification{
		Timestamp: time.Now().UnixNano(),
		Prefix:    r.Prefix,
		Update:    u,
	}
	//n, err := r.StateCache.GetNotificationFromJSON2(r.Prefix, &gnmi.Path{Elem: pe}, x, r.RootSchema)
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

func (r *node) DeleteStateCache() error {
	r.handleDelete()

	pe, err := r.GetPathElem(nil, true)
	if err != nil {
		return err
	}
	n := &gnmi.Notification{
		Timestamp: time.Now().UnixNano(),
		Prefix:    r.Prefix,
		Delete:    []*gnmi.Path{{Elem: pe}},
	}
	if err := r.StateCache.GnmiUpdate(r.Prefix.Target, n); err != nil {
		return err
	}
	return nil
}

func (r *node) GetData(resource string, key map[string]string) (interface{}, error) {
	switch {
	case resource == "endpoint":
		if n, ok := key["name"]; ok {
			return r.data.GetEndpoint(n), nil
		}
		return nil, errors.New(fmt.Sprintf("unexpected key: %v", key))
	default:
		return nil, errors.New(fmt.Sprintf("unexpected resource: %s", resource))
	}
}

func (r *node) SetData(resource string, key map[string]string, d interface{}) error {
	switch {
	case resource == "endpoint":
		ep, ok := d.(*topov1alpha1.NddotopologyTopologyNodeStateEndpoint)
		if !ok {
			return errors.New(fmt.Sprintf("unexpected data: %v", d))
		}
		r.data.SetEndpoint(ep)

		return r.UpdateStateCache()
	default:
		return errors.New(fmt.Sprintf("unexpected resource: %s", resource))
	}
}

func (r *node) GetTarget() *targetInfo {
	return r.parent.GetTarget()
}

func (r *node) GetTargets() []*types.TargetConfig {
	return r.parent.GetTargets()
}

func (r *node) handleDelete() {
	r.Log.Debug("handleDelete", "Data", r.data)
}

func (r *node) handleConfigUpdate() error {
	r.Log.Debug("handleConfigUpdate", "Data", r.data)

	// set status
	r.handleStatus(r.parent.data.GetStatus())

	// set platform
	r.setPlatform()

	// set last update time
	r.data.SetLastUpdate(time.Now().String())

	// update state cache
	if err := r.UpdateStateCache(); err != nil {
		r.Log.Debug("error updating state cache")
	}
	return nil
}

func (r *node) handleTopologyUpdate(parentStatus string) {
	r.Log.Debug("handleTopologyUpdate", "Data", r.data)

	// set status
	r.handleStatus(parentStatus)

	// set last update time
	r.data.SetLastUpdate(time.Now().String())

	// update state cache
	if err := r.UpdateStateCache(); err != nil {
		r.Log.Debug("error updating state cache")
	}
}

func (r *node) handleStatus(parentStatus string) {
	if r.data.State == nil {
		r.data.State = &topov1alpha1.NddotopologyTopologyNodeState{
			Endpoint: make([]*topov1alpha1.NddotopologyTopologyNodeStateEndpoint, 0),
		}
	}
	if parentStatus == "down" {
		r.data.SetStatus("down")
		r.data.SetReason("parent status down")
	} else {
		//r.Log.Debug("link", "status", r.data.GetStatus(), "admin-state", r.data.GetAdminState())
		// transition down -> up
		if (r.data.GetStatus() == "up" || r.data.GetStatus() == "") && r.data.GetAdminState() == "disable" {
			r.data.SetStatus("down")
			r.data.SetReason("admin disabled")
		}

		// transition up -> down
		if (r.data.GetStatus() == "down" || r.data.GetStatus() == "") && r.data.GetAdminState() == "enable" {
			r.data.SetStatus("up")
			r.data.SetReason("")
		}
	}
}

func (r *node) setPlatform() {
	if r.data.GetPlatform() == "" {
		// platform is not defined at node level
		for _, k := range r.parent.data.GetKinds() {
			if k.GetName() == r.data.GetKindName() {
				p := k.GetPlatform()
				if p != "" {
					r.data.SetPlatform(p)
					return
				}
			}
		}
		// platform is not defined at kind level
		d := r.parent.data.GetDefaults()
		if d != nil {
			p := d.GetPlatform()
			if p != "" {
				r.data.SetPlatform(p)
				return
			}
		}
		// platform is not defined we use the global default
		r.data.SetPlatform("ixrd2")
		return

	}
	// all good since the platform is already set
}
