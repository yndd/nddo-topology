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

type topology struct {
	dispatcher.Resource
	data   *topov1alpha1.NddotopologyTopology
	parent *root
	//activeTargets []*types.TargetConfig // active targets
	//target        *targetInfo           // observer interface
	//ctx           context.Context
	nodes map[string]dispatcher.Handler
	links map[string]dispatcher.Handler

	observerList []TopologyObserver
}

func (r *topology) WithLogging(log logging.Logger) {
	r.Log = log
}

func (r *topology) WithStateCache(c *cache.Cache) {
	r.StateCache = c
}

func (r *topology) WithTargetCache(c *cache.Cache) {
	r.TargetCache = c
}

func (r *topology) WithConfigCache(c *cache.Cache) {
	r.ConfigCache = c
}

func (r *topology) WithPrefix(p *gnmi.Path) {
	r.Prefix = p
}

func (r *topology) WithPathElem(pe []*gnmi.PathElem) {
	r.PathElem = pe[0]
}

func (r *topology) WithRootSchema(rs *yentry.Entry) {
	r.RootSchema = rs
}

func (r *topology) WithK8sClient(c client.Client) {
	r.Client = c
}

func NewTopology(n string, opts ...dispatcher.Option) dispatcher.Handler {
	x := &topology{
		nodes:        make(map[string]dispatcher.Handler),
		links:        make(map[string]dispatcher.Handler),
		observerList: []TopologyObserver{},
	}
	x.Key = n

	for _, opt := range opts {
		opt(x)
	}
	return x
}

func topologyGetKey(p []*gnmi.PathElem) string {
	return p[0].GetKey()["name"]
}

func topologyCreate(log logging.Logger, cc, sc, tc *cache.Cache, c client.Client, prefix *gnmi.Path, p []*gnmi.PathElem, d interface{}) dispatcher.Handler {
	topologyName := topologyGetKey(p)
	return NewTopology(topologyName,
		dispatcher.WithPrefix(prefix),
		dispatcher.WithPathElem(p),
		dispatcher.WithLogging(log),
		dispatcher.WithStateCache(sc),
		dispatcher.WithTargetCache(tc),
		dispatcher.WithConfigCache(cc),
		dispatcher.WithK8sClient(c))
}

func (r *topology) HandleConfigEvent(o dispatcher.Operation, prefix *gnmi.Path, pe []*gnmi.PathElem, d interface{}) (dispatcher.Handler, error) {
	log := r.Log.WithValues("Operation", o, "Path Elem", pe)

	log.Debug("HandleConfigEvent topology")

	children := map[string]dispatcher.HandleConfigEventFunc{
		"node": nodeCreate,
		"link": linkCreate,
	}

	// check path Element Name
	pathElemName := pe[0].GetName()
	if _, ok := children[pathElemName]; !ok {
		return nil, errors.Wrap(errors.New("unexpected pathElem"), fmt.Sprintf("topology HandleConfigEvent: %s", pathElemName))
	}

	if len(pe) == 1 {
		//log.Debug("topology Handle pathelem =1")
		// handle local
		switch o {
		case dispatcher.OperationUpdate:
			i, err := r.CreateChild(children, pathElemName, prefix, pe, d)
			if err != nil {
				return nil, err
			}
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
		//log.Debug("topology Handle pathelem >1")
		i, err := r.CreateChild(children, pathElemName, prefix, pe[:1], nil)
		if err != nil {
			return nil, err
		}
		return i.HandleConfigEvent(o, prefix, pe[1:], d)
	}
	return nil, nil

}

func (r *topology) CreateChild(children map[string]dispatcher.HandleConfigEventFunc, pathElemName string, prefix *gnmi.Path, pe []*gnmi.PathElem, d interface{}) (dispatcher.Handler, error) {
	switch pathElemName {
	case "node":
		if i, ok := r.nodes[nodeGetKey(pe)]; !ok {
			i = children[pathElemName](r.Log, r.ConfigCache, r.StateCache, r.TargetCache, r.Client, prefix, pe, d)
			i.SetRootSchema(r.RootSchema)
			if err := i.SetParent(r); err != nil {
				return nil, err
			}
			r.nodes[nodeGetKey(pe)] = i
			return i, nil
		} else {
			return i, nil
		}
	case "link":
		if i, ok := r.links[linkGetKey(pe)]; !ok {
			i = children[pathElemName](r.Log, r.ConfigCache, r.StateCache, r.TargetCache, r.Client, prefix, pe, d)
			i.SetRootSchema(r.RootSchema)
			if err := i.SetParent(r); err != nil {
				return nil, err
			}
			r.links[linkGetKey(pe)] = i
			return i, nil
		} else {
			return i, nil
		}
	}
	return nil, nil
}

func (r *topology) DeleteChild(pathElemName string, pe []*gnmi.PathElem) error {
	switch pathElemName {
	case "node":
		if i, ok := r.nodes[nodeGetKey(pe)]; ok {
			if err := i.DeleteStateCache(); err != nil {
				return err
			}
		}
	case "link":
		if i, ok := r.links[linkGetKey(pe)]; ok {
			if err := i.DeleteStateCache(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *topology) SetParent(parent interface{}) error {
	p, ok := parent.(*root)
	if !ok {
		return errors.New("wrong parent object")
	}
	r.parent = p
	return nil
}

func (r *topology) SetRootSchema(rs *yentry.Entry) {
	r.RootSchema = rs
}

func (r *topology) GetChildren() map[string]string {
	x := make(map[string]string)
	for k := range r.nodes {
		x[k] = "node"
	}
	for k := range r.links {
		x[k] = "link"
	}
	return x
}

func (r *topology) UpdateConfig(d interface{}) error {
	r.Copy(d)
	// from the moment we get dat we are interested in target updates
	//r.GetTarget().RegisterTargetObserver(r)
	// add channel here
	if err := r.handleConfigUpdate(); err != nil {
		return err
	}
	//r.data.SetLastUpdate(time.Now().String())
	// update the state cache
	if err := r.UpdateStateCache(); err != nil {
		return err
	}
	return nil
}

func (r *topology) GetPathElem(p []*gnmi.PathElem, do_recursive bool) ([]*gnmi.PathElem, error) {
	//r.Log.Debug("GetPathElem", "PathElem topology", r.PathElem)
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

func (r *topology) Copy(d interface{}) error {
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	x := topov1alpha1.NddotopologyTopology{}
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}
	r.data = (&x).DeepCopy()
	//r.Log.Debug("Copy", "Data", r.data)
	return nil
}

func (r *topology) UpdateStateCache() error {
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

func (r *topology) DeleteStateCache() error {
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

func (r *topology) GetData(resource string, key map[string]string) (interface{}, error) {
	return nil, nil
}

func (r *topology) SetData(resource string, key map[string]string, d interface{}) error {
	return nil
}

func (r *topology) GetTarget() *targetInfo {
	return r.parent.GetTarget()
}

func (r *topology) GetTargets() []*types.TargetConfig {
	return r.parent.GetTargets()
}

func (r *topology) handleDelete() error {
	r.Log.Debug("handleDelete", "Data", r.data)
	return nil
}

func (r *topology) handleConfigUpdate() error {
	r.Log.Debug("handleConfigUpdate", "Data", r.data)

	if r.data.State == nil {
		r.data.State = &topov1alpha1.NddotopologyTopologyState{}

	}
	if (r.data.GetStatus() == "up" || r.data.GetStatus() == "") && r.data.GetAdminState() == "disable" {
		r.data.SetStatus("down")
		r.data.SetReason("admin disabled")
		for _, o := range r.observerList {
			o.handleTopologyUpdate(r.data.GetStatus())
		}
		r.data.SetLastUpdate(time.Now().String())
	}

	if (r.data.GetStatus() == "down" || r.data.GetStatus() == "") && r.data.GetAdminState() == "enable" {
		r.data.SetStatus("up")
		r.data.SetReason("")

		for _, o := range r.observerList {
			o.handleTopologyUpdate(r.data.GetStatus())
		}
	}

	r.data.SetLastUpdate(time.Now().String())

	return nil
}

func (r *topology) RegisterTopologyObserver(o TopologyObserver) {
	r.observerList = append(r.observerList, o)
}

func (r *topology) RemoveTopologyObserver(o TopologyObserver) {
	found := false
	i := 0
	for ; i < len(r.observerList); i++ {
		if r.observerList[i] == o {
			found = true
			break
		}
	}
	if found {
		r.observerList = append(r.observerList[:i], r.observerList[i+1:]...)
	}
}

func (r *topology) NotifyTargetObserver() {
	for _, observer := range r.observerList {
		observer.handleTopologyUpdate(*r.data.State.Status)
	}
}

type TopologySubject interface {
	RegisterTopologyObserver(o TargetObserver)
	RemoveTopologyObserver(o TargetObserver)
	NotifyTopologyObserver()
}

type TopologyObserver interface {
	handleTopologyUpdate(string)
}
