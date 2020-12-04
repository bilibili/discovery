package discovery

import (
	"context"

	"github.com/bilibili/discovery/model"
	"github.com/bilibili/discovery/registry"
	"github.com/go-kratos/kratos/pkg/ecode"
	log "github.com/go-kratos/kratos/pkg/log"
)

// Register a new instance.
func (d *Discovery) Register(c context.Context, ins *model.Instance, latestTimestamp int64, replication bool, fromzone bool) {
	_ = d.registry.Register(ins, latestTimestamp)
	if !replication {
		_ = d.nodes.Load().(*registry.Nodes).Replicate(c, model.Register, ins, fromzone)
	}
}

// Renew marks the given instance of the given app name as renewed, and also marks whether it originated from replication.
func (d *Discovery) Renew(c context.Context, arg *model.ArgRenew) (i *model.Instance, err error) {
	i, ok := d.registry.Renew(arg)
	if !ok {
		err = ecode.NothingFound
		log.Error("renew appid(%s) hostname(%s) zone(%s) env(%s) error", arg.AppID, arg.Hostname, arg.Zone, arg.Env)
		return
	}
	if !arg.Replication {
		_ = d.nodes.Load().(*registry.Nodes).Replicate(c, model.Renew, i, arg.Zone != d.c.Env.Zone)
		return
	}
	if arg.DirtyTimestamp > i.DirtyTimestamp {
		err = ecode.NothingFound
	} else if arg.DirtyTimestamp < i.DirtyTimestamp {
		err = ecode.Conflict
	}
	return
}

// Cancel cancels the registration of an instance.
func (d *Discovery) Cancel(c context.Context, arg *model.ArgCancel) (err error) {
	i, ok := d.registry.Cancel(arg)
	if !ok {
		err = ecode.NothingFound
		log.Error("cancel appid(%s) hostname(%s) error", arg.AppID, arg.Hostname)
		return
	}
	if !arg.Replication {
		_ = d.nodes.Load().(*registry.Nodes).Replicate(c, model.Cancel, i, arg.Zone != d.c.Env.Zone)
	}
	return
}

// FetchAll fetch all instances of all the department.
func (d *Discovery) FetchAll(c context.Context) (im map[string][]*model.Instance) {
	return d.registry.FetchAll()
}

// Fetch fetch all instances by appid.
func (d *Discovery) Fetch(c context.Context, arg *model.ArgFetch) (info *model.InstanceInfo, err error) {
	return d.registry.Fetch(arg.Zone, arg.Env, arg.AppID, 0, arg.Status)
}

// Fetchs fetch multi app by appids.
func (d *Discovery) Fetchs(c context.Context, arg *model.ArgFetchs) (is map[string]*model.InstanceInfo, err error) {
	is = make(map[string]*model.InstanceInfo, len(arg.AppID))
	for _, appid := range arg.AppID {
		i, err := d.registry.Fetch(arg.Zone, arg.Env, appid, 0, arg.Status)
		if err != nil {
			log.Error("Fetchs fetch appid(%v) err", err)
			continue
		}
		is[appid] = i
	}
	return
}

// Polls hangs request and then write instances when that has changes, or return NotModified.
func (d *Discovery) Polls(c context.Context, arg *model.ArgPolls) (ch chan map[string]*model.InstanceInfo, new bool, miss []string, err error) {
	return d.registry.Polls(arg)
}

// DelConns delete conn of host in appid
func (d *Discovery) DelConns(arg *model.ArgPolls) {
	d.registry.DelConns(arg)
}

// Nodes get all nodes of discovery.
func (d *Discovery) Nodes(c context.Context) (nsi []*model.Node) {
	return d.nodes.Load().(*registry.Nodes).Nodes()
}

// Set set metadata,color,status of instance.
func (d *Discovery) Set(c context.Context, arg *model.ArgSet) (err error) {
	if !d.registry.Set(arg) {
		err = ecode.RequestErr
	}
	if !arg.Replication {
		d.nodes.Load().(*registry.Nodes).ReplicateSet(c, arg, arg.FromZone)
	}
	return
}
