package discovery

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/Bilibili/discovery/errors"
	"github.com/Bilibili/discovery/model"
	"github.com/Bilibili/discovery/registry"
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
			log.Errorf("d.client.Get(%v) error(%v)", uri, err)
			continue
		}
		if res.Code != 0 {
			log.Errorf("service syncup from(%s) failed ", uri)
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

func (d *Discovery) regSelf() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	now := time.Now().UnixNano()
	ins := &model.Instance{
		Region: os.Getenv("REGION"),
		Zone:   os.Getenv("ZONE"),
		Env:    os.Getenv("DEPLOY_ENV"),
		AppID:  model.AppID,
		Addrs: []string{
			"http://" + d.c.HTTPServer.Addr,
		},
		Status:          model.InstanceStatusUP,
		RegTimestamp:    now,
		UpTimestamp:     now,
		LatestTimestamp: now,
		RenewTimestamp:  now,
		DirtyTimestamp:  now,
	}
	ins.Hostname, _ = os.Hostname()
	d.Register(ctx, ins, now, false)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				arg := &model.ArgRenew{
					AppID:    ins.AppID,
					Zone:     ins.Zone,
					Env:      ins.Env,
					Hostname: ins.Hostname,
				}
				if _, err := d.Renew(ctx, arg); err != nil && err == errors.NothingFound {
					d.Register(ctx, ins, now, false)
				}
			case <-ctx.Done():
				arg := &model.ArgCancel{
					AppID:    model.AppID,
					Zone:     ins.Zone,
					Env:      ins.Env,
					Hostname: ins.Hostname,
				}
				if err := d.Cancel(context.Background(), arg); err != nil {
					log.Errorf("d.Cancel(%+v) error(%v)", arg, err)
				}
				return
			}
		}
	}()
	return cancel
}

func (d *Discovery) nodesproc() {
	var (
		lastTs int64
	)
	for {
		arg := &model.ArgPolls{
			AppID:           []string{model.AppID},
			Zone:            os.Getenv("ZONE"),
			Env:             os.Getenv("DEPLOY_ENV"),
			LatestTimestamp: []int64{lastTs},
		}
		arg.Hostname, _ = os.Hostname()
		ch, _, err := d.registry.Polls(arg)
		if err != nil && err != errors.NotModified {
			log.Errorf("d.registry(%v) error(%v)", arg, err)
			time.Sleep(time.Second)
			continue
		}
		apps := <-ch
		ins, ok := apps[model.AppID]
		if !ok || ins == nil {
			return
		}
		var (
			nodes []string
			zones = make(map[string]string)
		)
		for _, ins := range ins.Instances {
			for _, in := range ins {
				for _, addr := range in.Addrs {
					u, err := url.Parse(addr)
					if err == nil && u.Scheme == "http" {
						if in.Zone == arg.Zone {
							nodes = append(nodes, u.Host)
						} else {
							zones[u.Host] = in.Zone
						}
					}
				}
			}
		}
		lastTs = ins.LatestTimestamp
		log.Infof("discovery changed nodes:%v zones:%v", nodes, zones)
		d.c.Nodes = nodes
		d.c.Zones = zones
		d.nodes = registry.NewNodes(d.c)
		d.nodes.UP()
	}
}
