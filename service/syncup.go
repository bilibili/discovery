package service

import (
	"context"
	"fmt"

	log "golang/log4go"

	"github.com/Bilibili/discovery/model"
)

var (
	_fetchAllURL = "http://%s/discovery/fetch/all"
)

// syncUp populates the registry information from a peer eureka node.
func (s *Service) syncUp() (err error) {
	for _, node := range s.nodes.AllNodes() {
		if s.nodes.Myself(node.Addr) {
			continue
		}
		uri := fmt.Sprintf(_fetchAllURL, node.Addr)
		var res struct {
			Code int                          `json:"code"`
			Data map[string][]*model.Instance `json:"data"`
		}
		if err = s.client.Get(context.TODO(), uri, "", nil, &res); err != nil {
			log.Error("e.client.Get(%v) error(%v)", uri, err)
			continue
		}
		if res.Code != 0 {
			log.Error("service syncup from(%s) failed ", uri)
			continue
		}
		for _, is := range res.Data {
			for _, i := range is {
				s.tLock.RLock()
				appid, ok := s.tree[i.Treeid]
				s.tLock.RUnlock()
				if !ok || appid != i.Appid {
					s.tLock.Lock()
					s.tree[i.Treeid] = i.Appid
					s.tLock.Unlock()
				}
				s.registry.Register(i, i.LatestTimestamp)
			}
		}
		// NOTE: no return, make sure that all instances from other nodes register into self.
	}
	s.nodes.UP()
	return
}
