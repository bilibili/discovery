package registry

import (
	"context"
	"strings"
	"sync"

	"github.com/bilibili/discovery/model"
	"github.com/go-kratos/kratos/pkg/conf/paladin"
)

// Scheduler info.
type scheduler struct {
	schedulers map[string]*model.Scheduler
	mutex      sync.RWMutex
	r          *Registry
}

func newScheduler(r *Registry) *scheduler {
	return &scheduler{
		schedulers: make(map[string]*model.Scheduler),
		r:          r,
	}
}

// Load load scheduler info.
func (s *scheduler) Load() {
	for _, key := range paladin.Keys() {
		if !strings.HasSuffix(key, ".json") {
			continue
		}
		v := paladin.Get(key)
		content, err := v.String()
		if err != nil {
			return
		}
		sch := new(model.Scheduler)
		if err := sch.Set(content); err != nil {
			continue
		}
		s.schedulers[appsKey(sch.AppID, sch.Env)] = sch
	}
}

// Reload reload scheduler info.
func (s *scheduler) Reload() {
	event := paladin.WatchEvent(context.Background())
	for {
		e := <-event

		if !strings.HasSuffix(e.Key, ".json") {
			continue
		}
		sch := new(model.Scheduler)
		if err := sch.Set(e.Value); err != nil {
			continue
		}
		s.mutex.Lock()
		key := appsKey(sch.AppID, sch.Env)
		s.r.aLock.Lock()
		if a, ok := s.r.appm[key]; ok {
			a.UpdateLatest(0)
		}
		s.r.aLock.Unlock()
		s.schedulers[key] = sch
		s.mutex.Unlock()
		s.r.broadcast(sch.Env, sch.AppID)
	}
}

// Get get scheduler info.
func (s *scheduler) Get(appid, env string) *model.Scheduler {
	s.mutex.RLock()
	sch := s.schedulers[appsKey(appid, env)]
	s.mutex.RUnlock()
	return sch
}
