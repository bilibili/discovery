package model

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	log "golang/log4go"
)

// InstanceStatus Status of instance
// type InstanceStatus uint32

const (
	// InstanceStatusUP Ready to receive traffic
	InstanceStatusUP = uint32(1)
	// InstancestatusWating Intentionally shutdown for traffic
	InstancestatusWating = uint32(1) << 1
)

func (i *Instance) filter(status uint32) bool {
	return status&i.Status > 0
}

// Action Replicate type of node
type Action int

const (
	// Register Replicate the add action to all nodes
	Register Action = iota
	// Renew Replicate the heartbeat action to all nodes
	Renew
	// Cancel Replicate the cancel action to all nodes
	Cancel
	// Weight Replicate the Weight action to all nodes
	Weight
	// Delete Replicate the Delete action to all nodes
	Delete
	// Status Replicate the Status action to all nodes
	Status
	// Color Replicate the Status action to all nodes
	Color
)

// Instance holds information required for registration with
// <Discovery Server> and to be discovered by other components.
type Instance struct {
	Region   string          `json:"region"`
	Zone     string          `json:"zone"`
	Env      string          `json:"env"`
	Appid    string          `json:"appid"`
	Treeid   int64           `json:"treeid"`
	Hostname string          `json:"hostname"`
	HTTP     string          `json:"http"`
	RPC      string          `json:"rpc"`
	Color    string          `json:"color"`
	Weight   int             `json:"weight"`
	Version  string          `json:"version"`
	Metadata json.RawMessage `json:"metadata"`

	// Status enum instance status
	Status uint32 `json:"status"`

	// timestamp
	RegTimestamp   int64 `json:"reg_timestamp"`
	UpTimestamp    int64 `json:"up_timestamp"` // NOTE: It is latest timestamp that status becomes UP.
	RenewTimestamp int64 `json:"renew_timestamp"`
	DirtyTimestamp int64 `json:"dirty_timestamp"`

	LatestTimestamp int64 `json:"latest_timestamp"`
}

// NewInstance new a instance.
func NewInstance(arg *ArgRegister) (i *Instance) {
	now := time.Now().UnixNano()
	i = &Instance{
		Region:         arg.Region,
		Zone:           arg.Zone,
		Env:            arg.Env,
		Appid:          arg.Appid,
		Treeid:         arg.Treeid,
		Hostname:       arg.Hostname,
		HTTP:           arg.HTTP,
		RPC:            arg.RPC,
		Color:          arg.Color,
		Weight:         arg.Weight,
		Version:        arg.Version,
		Status:         arg.Status,
		RegTimestamp:   now,
		UpTimestamp:    now,
		RenewTimestamp: now,
		DirtyTimestamp: now,
	}
	return
}

// InstanceInfo the info get by consumer.
type InstanceInfo struct {
	Instances       []*Instance `json:"instances"`
	LatestTimestamp int64       `json:"latest_timestamp"`
}

// Department department distinguished by environment (example: shsb.prob)
type Department struct {
	apps map[string]map[string]*App // region.env->appid->hostname
	lock sync.RWMutex
}

// NewDepartment new Department.
func NewDepartment() (d *Department) {
	d = &Department{
		apps: make(map[string]map[string]*App),
	}
	return
}

// Apps returns all apps.
func (d *Department) Apps() (as []*App) {
	d.lock.RLock()
	for _, am := range d.apps {
		for _, a := range am {
			as = append(as, a)
		}
	}
	d.lock.RUnlock()
	return
}

// App find app by region and env.
func (d *Department) App(region, zone, env, appid string) (a *App, ok bool) {
	re := eKey(region, zone, env)
	d.lock.RLock()
	as, ok := d.apps[re]
	if ok {
		a, ok = as[appid]
	}
	d.lock.RUnlock()
	return
}

// NewApp news a app by appid. If ok=false, returns the app of already exist.
func (d *Department) NewApp(region, zone, env, appid string, treeid int64) (a *App, new bool) {
	re := eKey(region, zone, env)
	d.lock.Lock()
	as, ok := d.apps[re]
	if !ok {
		as = make(map[string]*App)
		d.apps[re] = as
	}
	if a, ok = as[appid]; !ok {
		a = NewApp(appid, treeid)
		as[appid] = a
	}
	d.lock.Unlock()
	new = !ok
	return
}

// eKey return region.env
func eKey(region, zone, env string) string {
	return fmt.Sprintf("%s.%s.%s", region, zone, env)
}

// Del deletes app by region and env.
func (d *Department) Del(region, zone, env, appid string) {
	re := eKey(region, zone, env)
	d.lock.Lock()
	defer d.lock.Unlock()
	as, ok := d.apps[re]
	if !ok {
		return
	}
	delete(as, appid)
	if len(as) == 0 {
		delete(d.apps, re)
	}
}

// App Instances distinguished by hostname
type App struct {
	ID              string
	Treeid          int64
	instances       map[string]*Instance
	latestTimestamp int64

	lock sync.RWMutex
}

// NewApp new App.
func NewApp(appid string, treeid int64) (a *App) {
	a = &App{
		ID:        appid,
		Treeid:    treeid,
		instances: make(map[string]*Instance),
	}
	return
}

// Instances return slice of instances.
func (a *App) Instances() (is []*Instance) {
	a.lock.RLock()
	is = make([]*Instance, 0, len(a.instances))
	for _, i := range a.instances {
		ni := new(Instance)
		*ni = *i
		is = append(is, ni)
	}
	a.lock.RUnlock()
	return
}

// InstanceInfo return slice of instances.if up is true,return all status instance else return up status instance
func (a *App) InstanceInfo(latestTime int64, status uint32) (ci *InstanceInfo, ok bool) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	if latestTime >= a.latestTimestamp {
		return
	}
	ci = &InstanceInfo{
		LatestTimestamp: a.latestTimestamp,
	}
	for _, i := range a.instances {
		// if up is false return all status instance
		if i.filter(status) {
			// if i.Status == InstanceStatusUP && i.LatestTimestamp > latestTime { // TODO(felix): increase
			ni := new(Instance)
			*ni = *i
			ci.Instances = append(ci.Instances, ni)
		}
	}
	ok = true
	return
}

// NewInstance new a instance.
func (a *App) NewInstance(ni *Instance, latestTime int64) (i *Instance, ok bool) {
	i = new(Instance)
	a.lock.Lock()
	oi, ok := a.instances[ni.Hostname]
	if ok {
		ni.UpTimestamp = oi.UpTimestamp
		if ni.DirtyTimestamp < oi.DirtyTimestamp {

			log.Warn("register exist(%v) dirty timestamp over than caller(%v)", oi, ni)
			ni = oi
		}
	}
	a.instances[ni.Hostname] = ni
	if latestTime == 0 {
		latestTime = time.Now().UnixNano()
	}
	ni.LatestTimestamp = latestTime
	if latestTime > a.latestTimestamp {
		a.latestTimestamp = latestTime
	}
	*i = *ni
	a.lock.Unlock()
	ok = !ok
	return
}

// Renew new a instance.
func (a *App) Renew(hostname string) (i *Instance, ok bool) {
	i = new(Instance)
	a.lock.Lock()
	defer a.lock.Unlock()
	oi, ok := a.instances[hostname]
	if !ok {
		return
	}
	oi.RenewTimestamp = time.Now().UnixNano()
	*i = *oi
	return
}

// Cancel cancel a instance.
func (a *App) Cancel(hostname string, latestTime int64) (i *Instance, l int, ok bool) {
	i = new(Instance)
	a.lock.Lock()
	defer a.lock.Unlock()
	oi, ok := a.instances[hostname]
	if !ok {
		return
	}
	delete(a.instances, hostname)
	l = len(a.instances)
	if latestTime == 0 {
		latestTime = time.Now().UnixNano()
	}
	oi.LatestTimestamp = latestTime
	a.latestTimestamp = latestTime
	*i = *oi
	return
}

// Len returns the length of instances.
func (a *App) Len() (l int) {
	a.lock.RLock()
	l = len(a.instances)
	a.lock.RUnlock()
	return
}

// SetStatus with a new status for instance.
func (a *App) SetStatus(changes map[string]uint32, setTime int64) (ok bool) {
	a.lock.Lock()
	defer a.lock.Unlock()
	var dst *Instance
	for hostname, s := range changes {
		if dst, ok = a.instances[hostname]; !ok {
			log.Error("SetWeight hostname(%s) not found", hostname)
			return
		}
		if s != InstanceStatusUP && s != InstancestatusWating {
			log.Error("SetWeight change status(%s) is error", s)
			ok = false
			return
		}
	}
	if setTime == 0 {
		setTime = time.Now().UnixNano()
	}
	for hostname, s := range changes {
		if dst, ok = a.instances[hostname]; !ok {
			log.Error("SetWeight hostname(%s) not found", hostname)
			return // NOTE: impossible~~~
		}
		dst.Status = s
		dst.LatestTimestamp = setTime
		dst.DirtyTimestamp = setTime
		if dst.Status == InstanceStatusUP {
			dst.UpTimestamp = setTime
		}
	}
	a.latestTimestamp = setTime
	return
}

// SetWeight with a new weight for instance.
func (a *App) SetWeight(changes map[string]int, setTime int64) (ok bool) {
	a.lock.Lock()
	defer a.lock.Unlock()
	var dst *Instance
	for hostname := range changes {
		if dst, ok = a.instances[hostname]; !ok {
			log.Error("SetWeight hostname(%s) not found", hostname)
			return
		}
	}
	if setTime == 0 {
		setTime = time.Now().UnixNano()
	}
	for hostname, weight := range changes {
		if dst, ok = a.instances[hostname]; !ok {
			log.Error("setWeight hostname(%s) not found", hostname)
			return // NOTE: impossible~~~
		}
		dst.Weight = weight
		dst.LatestTimestamp = setTime
		dst.DirtyTimestamp = setTime
	}
	a.latestTimestamp = setTime
	return
}

// SetColor with a new color for instance.
func (a *App) SetColor(changes map[string]string, setTime int64) (ok bool) {
	a.lock.Lock()
	defer a.lock.Unlock()
	var dst *Instance
	for hostname := range changes {
		if dst, ok = a.instances[hostname]; !ok {
			log.Error("SetWeight hostname(%s) not found", hostname)
			return
		}
	}
	if setTime == 0 {
		setTime = time.Now().UnixNano()
	}
	for hostname, color := range changes {
		if dst, ok = a.instances[hostname]; !ok {
			log.Error(" SetWeight hostname(%s) not found", hostname)
			return // NOTE: impossible~~~
		}
		dst.Color = color
		dst.LatestTimestamp = setTime
		dst.DirtyTimestamp = setTime
	}
	a.latestTimestamp = setTime
	return
}
