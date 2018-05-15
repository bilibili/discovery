package service

import (
	"context"

	log "golang/log4go"

	"github.com/Bilibili/discovery/conf"
	"github.com/Bilibili/discovery/dao"
	"github.com/Bilibili/discovery/errors"
	"github.com/Bilibili/discovery/model"
)

// Register a new instance.
func (s *Service) Register(c context.Context, ins *model.Instance, arg *model.ArgRegister) {
	s.registry.Register(ins, arg.LatestTimestamp)
	s.tLock.RLock()
	appid, ok := s.tree[ins.Treeid]
	s.tLock.RUnlock()
	if !ok || appid != ins.Appid {
		s.tLock.Lock()
		s.tree[ins.Treeid] = ins.Appid
		s.tLock.Unlock()
	}
	if !arg.Replication {
		s.nodes.Replicate(c, model.Register, ins, arg.Zone != s.env.Zone)
	}
}

// Renew marks the given instance of the given app name as renewed, and also marks whether it originated from replication.
func (s *Service) Renew(c context.Context, arg *model.ArgRenew) (i *model.Instance, err error) {
	i, ok := s.registry.Renew(arg)
	if !ok {
		log.Error("renew appid(%s) hostname(%s) zone(%s) env(%s) error", arg.Appid, arg.Hostname, arg.Zone, arg.Env)
		return
	}
	if !arg.Replication {
		s.nodes.Replicate(c, model.Renew, i, arg.Zone != s.env.Zone)
		return
	}
	if arg.DirtyTimestamp > i.DirtyTimestamp {
		err = errors.NothingFound
		return
	} else if arg.DirtyTimestamp < i.DirtyTimestamp {
		err = errors.Conflict
	}
	return
}

// Cancel cancels the registration of an instance.
func (s *Service) Cancel(c context.Context, arg *model.ArgCancel) (err error) {
	i, ok := s.registry.Cancel(arg)
	if !ok {
		err = errors.NothingFound
		log.Error("cancel appid(%s) hostname(%s) error", arg.Appid, arg.Hostname)
		return
	}
	if !arg.Replication {
		s.nodes.Replicate(c, model.Cancel, i, arg.Zone != s.env.Zone)
	}
	return
}

// FetchAll fetch all instances of all the department.
func (s *Service) FetchAll(c context.Context) (im map[string][]*model.Instance) {
	return s.registry.FetchAll()
}

// Fetch fetch all instances by appid.
func (s *Service) Fetch(c context.Context, arg *model.ArgFetch) (info *model.InstanceInfo, err error) {
	return s.registry.Fetch(arg.Zone, arg.Env, arg.Appid, 0, arg.Status)
}

// Polls hangs request and then write instances when that has changes, or return NotModified.
func (s *Service) Polls(c context.Context, arg *model.ArgPolls) (ch chan map[string]*model.InstanceInfo, new bool, err error) {
	return s.registry.Polls(arg)
}

// DelConns delete conn of host in appid
func (s *Service) DelConns(arg *model.ArgPolls) {
	s.registry.DelConns(arg)
}

// Nodes get all nodes of discovery.
func (s *Service) Nodes(c context.Context) (nsi []*model.Node) {
	return s.nodes.Nodes()
}

func (s *Service) updateNodes() {
	for {
		<-conf.ConfCh
		s.nodes = dao.NewNodes(s.c)
		s.nodes.UP()
	}
}

// PutChan put chan into pool.
func (s *Service) PutChan(ch chan map[string]*model.InstanceInfo) {
	s.registry.PutChan(ch)
}
