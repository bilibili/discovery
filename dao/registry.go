package dao

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	log "golang/log4go"

	"github.com/Bilibili/discovery/errors"
	"github.com/Bilibili/discovery/model"
)

const (
	_evictThreshold = int64(90 * time.Second)
	_evictCeiling   = int64(3600 * time.Second)
)

// Registry handles replication of all operations to peer Discovery nodes to keep them all in sync.
type Registry struct {
	dm    map[string]*model.Department // department->env->appid->hostname
	dLock sync.RWMutex

	conns map[string]map[string]*conn // region.zone.env.appid-> host
	pool  sync.Pool
	cLock sync.RWMutex
	gd    *Guard
}

// conn the poll chan contains consumer.
type conn struct {
	ch         chan map[int64]*model.InstanceInfo // TODO(felix): increase
	latestTime int64
}

// newConn new consumer chan.
func newConn(ch chan map[int64]*model.InstanceInfo, latestTime int64) *conn {
	return &conn{ch: ch, latestTime: latestTime}
}

// NewRegistry new register.
func NewRegistry() (r *Registry) {
	r = &Registry{
		dm:    make(map[string]*model.Department),
		conns: make(map[string]map[string]*conn),
		gd:    new(Guard),
		pool:  sync.Pool{New: func() interface{} { return make(chan map[int64]*model.InstanceInfo, 1) }},
	}
	go r.proc()
	return
}

func (r *Registry) newDepartment(appid string) (d *model.Department, ok bool) {
	dep := dep(appid)
	r.dLock.Lock()
	if d, ok = r.dm[dep]; !ok {
		d = model.NewDepartment()
		r.dm[dep] = d
	}
	r.dLock.Unlock()
	return
}

func (r *Registry) department(appid string) (d *model.Department, ok bool) {
	dep := dep(appid)
	r.dLock.RLock()
	d, ok = r.dm[dep]
	r.dLock.RUnlock()
	return
}

// get the department of appid
func dep(appid string) (dep string) {
	dep = appid
	if strings.Count(appid, ".") == 2 {
		dep = appid[:strings.LastIndexByte(appid, '.')]
	}
	return
}

func (r *Registry) newApp(ins *model.Instance) (a *model.App) {
	d, _ := r.newDepartment(ins.Appid)
	a, _ = d.NewApp(ins.Region, ins.Zone, ins.Env, ins.Appid, ins.Treeid)
	return
}

func (r *Registry) app(region, zone, env, appid string) (a *model.App, d *model.Department, ok bool) {
	if d, ok = r.department(appid); !ok {
		return
	}
	a, ok = d.App(region, zone, env, appid)
	return
}

// Register a new instance.
func (r *Registry) Register(ins *model.Instance, latestTime int64) (err error) {
	a := r.newApp(ins)
	i, ok := a.NewInstance(ins, latestTime)
	if ok {
		r.gd.incrExp()
	}
	// NOTE: make sure free poll before update appid latest timestamp.
	r.broadcast(i.Region, i.Zone, i.Env, i.Appid, a)
	return
}

// Renew marks the given instance of the given app name as renewed, and also marks whether it originated from replication.
func (r *Registry) Renew(arg *model.ArgRenew) (i *model.Instance, ok bool) {
	a, _, ok := r.app(arg.Region, arg.Zone, arg.Env, arg.Appid)
	if !ok {
		return
	}
	if i, ok = a.Renew(arg.Hostname); !ok {
		return
	}
	r.gd.incrFac()
	return
}

// Cancel cancels the registration of an instance.
func (r *Registry) Cancel(arg *model.ArgCancel) (i *model.Instance, ok bool) {
	if i, ok = r.cancel(arg.Region, arg.Zone, arg.Env, arg.Appid, arg.Hostname, arg.LatestTimestamp); !ok {
		return
	}
	r.gd.decrExp()
	return
}

func (r *Registry) cancel(region, zone, env, appid, hostname string, latestTime int64) (i *model.Instance, ok bool) {
	var l int
	a, d, ok := r.app(region, zone, env, appid)
	if !ok {
		return
	}
	if i, l, ok = a.Cancel(hostname, latestTime); !ok {
		return
	}
	if l == 0 {
		r.dLock.Lock()
		if a.Len() == 0 {
			d.Del(region, zone, env, appid)
		}
		r.dLock.Unlock()
	}
	r.broadcast(region, zone, env, appid, a) // NOTE: make sure free poll before update appid latest timestamp.
	return
}

// FetchAll fetch all instances of all the families.
func (r *Registry) FetchAll() (im map[string][]*model.Instance) {
	im = make(map[string][]*model.Instance)
	for _, d := range r.registry() {
		for _, a := range d.Apps() {
			im[a.ID] = a.Instances()
		}
	}
	return
}

// return all instances
func (r *Registry) registry() (ds []*model.Department) {
	r.dLock.RLock()
	ds = make([]*model.Department, 0, len(r.dm))
	for _, d := range r.dm {
		ds = append(ds, d)
	}
	r.dLock.RUnlock()
	return ds
}

// Fetch fetch all instances by appid.
func (r *Registry) Fetch(region, zone, env, appid string, latestTime int64, status uint32) (info *model.InstanceInfo, err error) {
	a, _, ok := r.app(region, zone, env, appid)
	if !ok {
		err = errors.NothingFound
		return
	}
	if info, ok = a.InstanceInfo(latestTime, status); !ok {
		err = errors.NotModified
	}
	return
}

// Polls hangs request and then write instances when that has changes, or return NotModified.
func (r *Registry) Polls(arg *model.ArgPolls) (ch chan map[int64]*model.InstanceInfo, new bool, err error) {
	var (
		ins = make(map[int64]*model.InstanceInfo, len(arg.Treeid))
		in  *model.InstanceInfo
	)
	if len(arg.Appid) != len(arg.LatestTimestamp) {
		arg.LatestTimestamp = make([]int64, len(arg.Appid))
	}
	ch = r.pool.Get().(chan map[int64]*model.InstanceInfo)
	for i := range arg.Appid {
		in, err = r.Fetch(arg.Region, arg.Zone, arg.Env, arg.Appid[i], arg.LatestTimestamp[i], model.InstanceStatusUP)
		if err == errors.NothingFound {
			log.Error("Polls region(%s) zone(%s) env(%s) appid(%s) error(%v)", arg.Region, arg.Zone, arg.Env, arg.Appid[i], err)
			return
		}
		if err == nil {
			ins[arg.Treeid[i]] = in
			new = true
		}
	}
	if new {
		ch <- ins
		return
	}
	r.cLock.Lock()
	ch = r.pool.Get().(chan map[int64]*model.InstanceInfo)
	for i := range arg.Appid {
		k := pollKey(arg.Region, arg.Zone, arg.Env, arg.Appid[i])
		connection := newConn(ch, arg.LatestTimestamp[i])
		if _, ok := r.conns[k]; !ok {
			r.conns[k] = make(map[string]*conn, 1)
		}
		r.conns[k][arg.Hostname] = connection
	}
	r.cLock.Unlock()
	return
}

// broadcast on poll by chan.
// NOTE: make sure free poll before update appid latest timestamp.
func (r *Registry) broadcast(region, zone, env, appid string, a *model.App) {
	key := pollKey(region, zone, env, appid)
	r.cLock.Lock()
	conns, ok := r.conns[key]
	if !ok {
		r.cLock.Unlock()
		return
	}
	delete(r.conns, key)
	r.cLock.Unlock()
	ci, _ := a.InstanceInfo(0, model.InstanceStatusUP) // TODO(felix): increase
	for _, conn := range conns {
		// c, ok := a.Consumer(cch.LatestTime) // TODO(felix): increase
		select {
		case conn.ch <- map[int64]*model.InstanceInfo{a.Treeid: ci}: // NOTE: if chan is full, means no poller.
		default:
			log.Error("s.broadcast err, channel full(%v)", ci)
		}
	}
}

func pollKey(region, zone, env, appid string) string {
	return fmt.Sprintf("%s.%s.%s.%s", region, zone, env, appid)
}

// Set Set the status of instance by hostnames.
func (r *Registry) Set(action model.Action, region, zone, env, appid string, changes map[string]string, setTime int64) (ok bool) {
	a, _, ok := r.app(region, zone, env, appid)
	if !ok {
		return
	}
	switch action {
	case model.Status:
		cs := make(map[string]uint32)
		for k, v := range changes {
			tmp, _ := strconv.ParseUint(v, 10, 32)
			cs[k] = uint32(tmp)
		}
		if ok = a.SetStatus(cs, setTime); !ok {
			return
		}
	case model.Weight:
		cs := make(map[string]int)
		for k, v := range changes {
			cs[k], _ = strconv.Atoi(v)
		}
		if ok = a.SetWeight(cs, setTime); !ok {
			return
		}
	case model.Color:
		if ok = a.SetColor(changes, setTime); !ok {
			return
		}
	}
	r.broadcast(region, zone, env, appid, a)
	return
}

// reset expect renews, count the renew of all app, one app has two expect remews in minute.
func (r *Registry) resetExp() {
	cnt := int64(0)
	for _, p := range r.registry() {
		for _, a := range p.Apps() {
			cnt += int64(a.Len())
		}
	}
	r.gd.setExp(cnt)
}

func (r *Registry) proc() {
	tk := time.Tick(1 * time.Minute)
	tk2 := time.Tick(15 * time.Minute)
	for {
		select {
		case <-tk:
			r.gd.updateFac()
			r.evict()
		case <-tk2:
			r.resetExp()
		}
	}
}

func (r *Registry) evict() {
	protect := r.gd.ok()
	// We collect first all expired items, to evict them in random order. For large eviction sets,
	// if we do not that, we might wipe out whole apps before self preservation kicks in. By randomizing it,
	// the impact should be evenly distributed across all applications.
	var eis []*model.Instance
	var registrySize int
	// all projects
	ps := r.registry()
	for _, p := range ps {
		for _, a := range p.Apps() {
			registrySize += a.Len()
			is := a.Instances()
			for _, i := range is {
				delta := time.Now().UnixNano() - i.RenewTimestamp
				if (!protect && delta > _evictThreshold) || delta > _evictCeiling {
					eis = append(eis, i)
				}
			}
		}
	}
	// To compensate for GC pauses or drifting local time, we need to use current registry size as a base for
	// triggering self-preservation. Without that we would wipe out full registry.
	eCnt := len(eis)
	registrySizeThreshold := int(float64(registrySize) * _percentThreshold)
	evictionLimit := registrySize - registrySizeThreshold
	if eCnt > evictionLimit {
		eCnt = evictionLimit
	}
	if eCnt == 0 {
		return
	}
	for i := 0; i < eCnt; i++ {
		// Pick a random item (Knuth shuffle algorithm)
		next := i + rand.Intn(len(eis)-i)
		eis[i], eis[next] = eis[next], eis[i]
		ei := eis[i]
		r.cancel(ei.Region, ei.Zone, ei.Env, ei.Appid, ei.Hostname, ei.LatestTimestamp)
	}
}

// DelConns delete conn of host in appid
func (r *Registry) DelConns(arg *model.ArgPolls) {
	var connection *conn
	r.cLock.Lock()
	for i := range arg.Appid {
		k := pollKey(arg.Region, arg.Zone, arg.Env, arg.Appid[i])
		conns, ok := r.conns[k]
		if !ok {
			log.Warn("DelConn key(%s) not found", k)
			continue
		}
		conn, ok := conns[arg.Hostname]
		if !ok {
			log.Warn("DelConn key(%s) host(%s) not found", k, arg.Hostname)
			continue
		}
		connection = conn
		delete(conns, arg.Hostname)
	}
	r.cLock.Unlock()
	if connection != nil {
		r.PutChan(connection.ch)
	}
}

// PutChan put chan into pool.
func (r *Registry) PutChan(ch chan map[int64]*model.InstanceInfo) {
	r.pool.Put(ch)
}
