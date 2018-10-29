package registry

import (
	"context"
	"fmt"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"github.com/Bilibili/discovery/conf"
	"github.com/Bilibili/discovery/model"
)

// Nodes is helper to manage lifecycle of a collection of Nodes.
type Nodes struct {
	c        *conf.Config
	nodes    atomic.Value
	selfAddr string
}

// NewNodes new nodes and return.
func NewNodes(c *conf.Config) (nodes *Nodes) {
	nodes = &Nodes{
		c:        c,
		selfAddr: c.HTTPServer.Addr,
	}
	nodes.Update(c.Nodes, c.Zones, 0)
	return
}

// Update ndoes and zones.
func (ns *Nodes) Update(nodes []string, zones map[string]string, status model.NodeStatus) {
	newNodes := make([]*Node, 0, len(ns.c.Nodes))
	for _, addr := range nodes {
		n := newNode(ns.c, addr)
		n.zone = ns.c.Zone
		n.pRegisterURL = fmt.Sprintf("http://%s%s", ns.c.HTTPServer.Addr, _registerURL)
		if ns.Myself(addr) {
			n.status = status
		}
		newNodes = append(newNodes, n)
	}
	for addr, name := range zones {
		n := newNode(ns.c, addr)
		n.zone = name
		n.otherZone = true
		n.pRegisterURL = fmt.Sprintf("http://%s%s", ns.c.HTTPServer.Addr, _registerURL)
		newNodes = append(newNodes, n)
	}
	ns.nodes.Store(newNodes)
}

// Replicate replicate information to all nodes except for this node.
func (ns *Nodes) Replicate(c context.Context, action model.Action, i *model.Instance, otherZone bool) (err error) {
	nodes := ns.nodes.Load().([]*Node)
	if len(nodes) == 0 {
		return
	}
	eg, c := errgroup.WithContext(c)
	for _, n := range nodes {
		if !ns.Myself(n.addr) {
			if otherZone && n.otherZone {
				continue
			}
			ns.action(c, eg, action, n, i)
		}
	}
	err = eg.Wait()
	return
}

func (ns *Nodes) action(c context.Context, eg *errgroup.Group, action model.Action, n *Node, i *model.Instance) {
	switch action {
	case model.Register:
		eg.Go(func() error {
			n.Register(c, i)
			return nil
		})
	case model.Renew:
		eg.Go(func() error {
			n.Renew(c, i)
			return nil
		})
	case model.Cancel:
		eg.Go(func() error {
			n.Cancel(c, i)
			return nil
		})
	}
}

// Nodes returns nodes of local zone.
func (ns *Nodes) Nodes() (nsi []*model.Node) {
	nodes := ns.nodes.Load().([]*Node)
	nsi = make([]*model.Node, 0, len(nodes))
	for _, nd := range nodes {
		if nd.otherZone {
			continue
		}
		node := &model.Node{
			Addr:   nd.addr,
			Status: nd.status,
			Zone:   nd.zone,
		}
		nsi = append(nsi, node)
	}
	return
}

// AllNodes returns nodes contain other zone nodes.
func (ns *Nodes) AllNodes() (nsi []*model.Node) {
	nodes := ns.nodes.Load().([]*Node)
	nsi = make([]*model.Node, 0, len(nodes))
	for _, nd := range nodes {
		node := &model.Node{
			Addr:   nd.addr,
			Status: nd.status,
			Zone:   nd.zone,
		}
		nsi = append(nsi, node)
	}
	return
}

// Myself returns whether or not myself.
func (ns *Nodes) Myself(addr string) bool {
	return ns.selfAddr == addr
}

// UP marks status of myself node up.
func (ns *Nodes) UP() {
	nodes := ns.nodes.Load().([]*Node)
	for _, nd := range nodes {
		if ns.Myself(nd.addr) {
			nd.status = model.NodeStatusUP
		}
	}
}
