package dao

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/Bilibili/discovery/conf"
	"github.com/Bilibili/discovery/model"
)

// Nodes is helper to manage lifecycle of a collection of Nodes.
type Nodes struct {
	nodes    []*Node
	selfAddr string
}

// NewNodes new nodes and return.
func NewNodes(c *conf.Config) (nodes *Nodes) {
	ns := make([]*Node, 0, len(c.Nodes))
	for _, addr := range c.Nodes {
		n := newNode(c, addr)
		n.pRegisterURL = fmt.Sprintf("http://%s%s", c.HTTPServer.Addr, _registerURL)
		ns = append(ns, n)
	}
	for addr, name := range c.Zones {
		n := newNode(c, addr)
		n.otherZone = true
		n.zone = name
		n.pRegisterURL = fmt.Sprintf("http://%s%s", c.HTTPServer.Addr, _registerURL)
		ns = append(ns, n)
	}
	nodes = &Nodes{
		nodes:    ns,
		selfAddr: c.HTTPServer.Addr,
	}
	return
}

// Replicate replicate information to all nodes except for this node.
func (ns *Nodes) Replicate(c context.Context, action model.Action, i *model.Instance, otherZone bool) (err error) {
	if len(ns.nodes) == 0 {
		return
	}
	eg, c := errgroup.WithContext(c)
	for _, n := range ns.nodes {
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
	nsi = make([]*model.Node, 0, len(ns.nodes))
	for _, nd := range ns.nodes {
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
	nsi = make([]*model.Node, 0, len(ns.nodes))
	for _, nd := range ns.nodes {
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
	for _, nd := range ns.nodes {
		if ns.Myself(nd.addr) {
			nd.status = model.NodeStatusUP
		}
	}
}
