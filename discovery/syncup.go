package discovery

import (
	"context"
	"fmt"

	"github.com/Bilibili/discovery/model"
	log "github.com/golang/glog"
)

var (
	_fetchAllURL = "http://%s/discovery/fetch/all"
)

// syncUp populates the registry information from a peer eureka node.
func (d *Discovery) syncUp() (err error) {
	for _, node := range d.nodes.AllNodes() {
		if d.nodes.Myself(node.Addr) {
			continue
		}
		uri := fmt.Sprintf(_fetchAllURL, node.Addr)
		var res struct {
			Code int                          `json:"code"`
			Data map[string][]*model.Instance `json:"data"`
		}
		if err = d.client.Get(context.TODO(), uri, "", nil, &res); err != nil {
			log.Error("e.client.Get(%v) error(%v)", uri, err)
			continue
		}
		if res.Code != 0 {
			log.Error("service syncup from(%s) failed ", uri)
			continue
		}
		for _, is := range res.Data {
			for _, i := range is {
				d.registry.Register(i, i.LatestTimestamp)
			}
		}
		// NOTE: no return, make sure that all instances from other nodes register into self.
	}
	d.nodes.UP()
	return
}
