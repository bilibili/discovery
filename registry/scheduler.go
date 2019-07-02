package registry

import (
	"encoding/json"
	"sync"

	"github.com/bilibili/discovery/model"

	log "github.com/bilibili/kratos/pkg/log"
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
func (s *scheduler) Load(conf []byte) {
	if len(conf) == 0 {
		return
	}
	schs := make([]*model.Scheduler, 0)
	err := json.Unmarshal(conf, &schs)
	if err != nil {
		log.Error("load scheduler info err %v", err)
	}
	for _, sch := range schs {
		s.schedulers[appsKey(sch.AppID, sch.Env)] = sch
	}
}

// TODO:dynamic reload scheduler config.
// func (s *scheduler)Reolad(){
//
//}

// Get get scheduler info.
func (s *scheduler) Get(appid, env string) *model.Scheduler {
	s.mutex.RLock()
	sch := s.schedulers[appsKey(appid, env)]
	s.mutex.RUnlock()
	return sch
}
