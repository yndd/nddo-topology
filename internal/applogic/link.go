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
	"github.com/yndd/ndd-runtime/pkg/utils"
	"github.com/yndd/ndd-yang/pkg/cache"
	"github.com/yndd/ndd-yang/pkg/dispatcher"
	"github.com/yndd/ndd-yang/pkg/yentry"
	"github.com/yndd/ndd-yang/pkg/yparser"
	topov1alpha1 "github.com/yndd/nddo-topology/apis/topo/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type link struct {
	dispatcher.Resource
	data   *topov1alpha1.NddotopologyTopologyLink
	parent *topology
	//activeTargets []*types.TargetConfig // active targets
	//target        *targetInfo           // observer interface
	//ctx           context.Context
}

func (r *link) WithLogging(log logging.Logger) {
	r.Log = log
}

func (r *link) WithStateCache(c *cache.Cache) {
	r.StateCache = c
}

func (r *link) WithTargetCache(c *cache.Cache) {
	r.TargetCache = c
}

func (r *link) WithConfigCache(c *cache.Cache) {
	r.ConfigCache = c
}

func (r *link) WithPrefix(p *gnmi.Path) {
	r.Prefix = p
}

func (r *link) WithPathElem(pe []*gnmi.PathElem) {
	r.PathElem = pe[0]
}

func (r *link) WithRootSchema(rs *yentry.Entry) {
	r.RootSchema = rs
}

func (r *link) WithK8sClient(c client.Client) {
	r.Client = c
}

func NewLink(n string, opts ...dispatcher.Option) dispatcher.Handler {
	x := &link{}

	x.Key = n

	for _, opt := range opts {
		opt(x)
	}
	return x
}

func linkGetKey(p []*gnmi.PathElem) string {
	return p[0].GetKey()["name"]
}

func linkCreate(log logging.Logger, cc, sc, tc *cache.Cache, c client.Client, prefix *gnmi.Path, p []*gnmi.PathElem, d interface{}) dispatcher.Handler {
	linkName := linkGetKey(p)
	return NewLink(linkName,
		dispatcher.WithPrefix(prefix),
		dispatcher.WithPathElem(p),
		dispatcher.WithLogging(log),
		dispatcher.WithStateCache(sc),
		dispatcher.WithTargetCache(tc),
		dispatcher.WithConfigCache(cc),
		dispatcher.WithK8sClient(c))
}

func (r *link) HandleConfigEvent(o dispatcher.Operation, prefix *gnmi.Path, pe []*gnmi.PathElem, d interface{}) (dispatcher.Handler, error) {
	log := r.Log.WithValues("Operation", o, "Path Elem", pe)

	log.Debug("HandleConfigEvent link")

	if len(pe) == 1 {
		return nil, errors.New("the handle should have been terminated in the parent")
	} else {
		return nil, errors.New("there is no children in the link resource ")
	}
}

func (r *link) CreateChild(children map[string]dispatcher.HandleConfigEventFunc, pathElemName string, prefix *gnmi.Path, pe []*gnmi.PathElem, d interface{}) (dispatcher.Handler, error) {
	return nil, errors.New("there is no children in the link resource ")
}

func (r *link) DeleteChild(pathElemName string, pe []*gnmi.PathElem) error {
	return errors.New("there is no children in the link resource ")
}

func (r *link) SetParent(parent interface{}) error {
	p, ok := parent.(*topology)
	if !ok {
		return errors.New("wrong parent object")
	}
	r.parent = p

	r.parent.RegisterTopologyObserver(r)
	return nil
}

func (r *link) SetRootSchema(rs *yentry.Entry) {
	r.RootSchema = rs
}

func (r *link) GetChildren() map[string]string {
	x := make(map[string]string)
	return x
}

func (r *link) UpdateConfig(d interface{}) error {
	r.Copy(d)
	// from the moment we get dat we are interested in target updates
	//r.GetTarget().RegisterTargetObserver(r)
	// add channel here
	if err := r.handleConfigUpdate(); err != nil {
		return err
	}
	return nil
}

func (r *link) GetPathElem(p []*gnmi.PathElem, do_recursive bool) ([]*gnmi.PathElem, error) {
	//r.Log.Debug("GetPathElem", "PathElem link", r.PathElem)
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

func (r *link) Copy(d interface{}) error {
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	x := topov1alpha1.NddotopologyTopologyLink{}
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}
	r.data = (&x).DeepCopy()
	//r.Log.Debug("Copy", "Data", r.data)
	return nil
}

func (r *link) UpdateStateCache() error {
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
	/*
		for _, upd := range u {
			r.Log.Debug("Link Update", "Path", yparser.GnmiPath2XPath(upd.Path, true), "Value", upd.GetVal())
		}
	*/
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

func (r *link) DeleteStateCache() error {
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

func (r *link) GetData(resource string, key map[string]string) (interface{}, error) {
	return nil, nil
}

func (r *link) SetData(resource string, key map[string]string, d interface{}) error {
	return nil
}

func (r *link) GetTarget() *targetInfo {
	return r.parent.GetTarget()
}

func (r *link) GetTargets() []*types.TargetConfig {
	return r.parent.GetTargets()
}

func (r *link) handleDelete() error {
	r.Log.Debug("handleDelete", "Data", r.data)
	return nil
}

func (r *link) handleConfigUpdate() error {
	r.Log.Debug("handleConfigUpdate", "Data", r.data)

	// set status
	r.handleStatus(r.parent.data.GetStatus())

	// parse link parses the link and sets the endpoints of the node
	if err := r.parseLink(); err != nil {
		return err
	}

	// set last update time
	r.data.SetLastUpdate(time.Now().String())
	if err := r.UpdateStateCache(); err != nil {
		return err
	}
	return nil
}

func (r *link) handleTopologyUpdate(parentStatus string) {
	r.Log.Debug("handleTopologyUpdate", "Data", r.data)

	// set status
	r.handleStatus(parentStatus)

	// set last update time
	r.data.SetLastUpdate(time.Now().String())

	if err := r.UpdateStateCache(); err != nil {
		r.Log.Debug("error updating state cache")
	}
}

func (r *link) handleStatus(parentStatus string) {
	if r.data.State == nil {
		r.data.State = &topov1alpha1.NddotopologyTopologyLinkState{}
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

func (r *link) parseLink() error {
	//parse link
	nameNodeA := r.data.Endpoints[0].GetNodeName()
	if _, ok := r.parent.nodes[nameNodeA]; !ok {
		r.data.SetStatus("down")
		r.data.SetReason(fmt.Sprintf("link endpoint node %s not found in topology definition", nameNodeA))
		return errors.New(fmt.Sprintf("link endpoint node %s not found in topology definition", nameNodeA))
	}
	nodeA := r.parent.nodes[nameNodeA]

	nameNodeB := r.data.Endpoints[1].GetNodeName()
	if _, ok := r.parent.nodes[nameNodeB]; !ok {
		r.data.SetStatus("down")
		r.data.SetReason(fmt.Sprintf("link endpoint node %s not found in topology definition", nameNodeB))
		return errors.New(fmt.Sprintf("link endpoint node %s not found in topology definition", nameNodeB))
	}
	nodeB := r.parent.nodes[nameNodeB]

	// check if link is part of a lag
	if r.data.GetLag() {
		// check if lag was already present, it could have been created by a previous link
		lagNameA := r.data.Endpoints[0].GetLagName()
		lagNameB := r.data.Endpoints[1].GetLagName()

		nodeEpA, err := r.getNodeEndpoint(nodeA, lagNameA)
		if err != nil {
			return err
		}
		nodeEpB, err := r.getNodeEndpoint(nodeB, lagNameB)
		if err != nil {
			return err
		}
		//r.Log.Debug("lagName", "lagNameA", lagNameA, "lagNameB", lagNameB, "nodeEpA", nodeEpA, "nodeEpB", nodeEpB)
		found := false
		if nodeEpA != nil && nodeEpB != nil {
			found = true
		}
		//r.Log.Debug("lagName", "Found", found)
		// create the logical lag link, lag = true, but lagSubLink is false
		if !found {
			epA := &topov1alpha1.NddotopologyTopologyNodeStateEndpoint{
				Name:       utils.StringPtr(lagNameA),
				Lag:        utils.BoolPtr(true),
				LagSubLink: utils.BoolPtr(false),
			}
			epB := &topov1alpha1.NddotopologyTopologyNodeStateEndpoint{
				Name:       utils.StringPtr(lagNameB),
				Lag:        utils.BoolPtr(true),
				LagSubLink: utils.BoolPtr(false),
			}
			if err := r.setNodeEndpoint(nodeA, epA); err != nil {
				return err
			}
			if err := r.setNodeEndpoint(nodeB, epB); err != nil {
				return err
			}
		}
		// create the physical sub link of the lag
		epA := &topov1alpha1.NddotopologyTopologyNodeStateEndpoint{
			Name:       utils.StringPtr(r.data.Endpoints[0].GetInterfaceName()),
			Lag:        utils.BoolPtr(false),
			LagSubLink: utils.BoolPtr(true),
		}
		epB := &topov1alpha1.NddotopologyTopologyNodeStateEndpoint{
			Name:       utils.StringPtr(r.data.Endpoints[1].GetInterfaceName()),
			Lag:        utils.BoolPtr(false),
			LagSubLink: utils.BoolPtr(true),
		}
		if err := r.setNodeEndpoint(nodeA, epA); err != nil {
			return err
		}
		if err := r.setNodeEndpoint(nodeB, epB); err != nil {
			return err
		}

		return nil
	}

	// non lag link
	epA := &topov1alpha1.NddotopologyTopologyNodeStateEndpoint{
		Name:       utils.StringPtr(r.data.Endpoints[0].GetInterfaceName()),
		Lag:        utils.BoolPtr(false),
		LagSubLink: utils.BoolPtr(false),
	}
	epB := &topov1alpha1.NddotopologyTopologyNodeStateEndpoint{
		Name:       utils.StringPtr(r.data.Endpoints[1].GetInterfaceName()),
		Lag:        utils.BoolPtr(false),
		LagSubLink: utils.BoolPtr(false),
	}
	if err := r.setNodeEndpoint(nodeA, epA); err != nil {
		return err
	}
	if err := r.setNodeEndpoint(nodeB, epB); err != nil {
		return err
	}

	return nil
}

func (r *link) getNodeEndpoint(d dispatcher.Handler, n string) (*topov1alpha1.NddotopologyTopologyNodeStateEndpoint, error) {
	data, err := d.GetData("endpoint", map[string]string{"name": n})
	if err != nil {
		return nil, err
	}
	switch x := data.(type) {
	case *topov1alpha1.NddotopologyTopologyNodeStateEndpoint:
		return x, nil
	default:
		return nil, nil
	}
}

func (r *link) setNodeEndpoint(d dispatcher.Handler, ep *topov1alpha1.NddotopologyTopologyNodeStateEndpoint) error {
	return d.SetData("endpoint", map[string]string{"name": ep.GetName()}, ep)
}
