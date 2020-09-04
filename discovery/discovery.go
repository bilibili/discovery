/*
 * @Author: gitsrc
 * @Date: 2020-09-04 09:57:29
 * @LastEditors: gitsrc
 * @LastEditTime: 2020-09-04 10:05:14
 * @FilePath: /discovery/discovery/discovery.go
 */
package discovery

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bilibili/discovery/conf"
	"github.com/bilibili/discovery/registry"
	http "github.com/bilibili/kratos/pkg/net/http/blademaster"
)

// Discovery discovery.
type Discovery struct {
	c         *conf.Config
	protected bool
	client    *http.Client
	registry  *registry.Registry
	nodes     atomic.Value
	sync.RWMutex
}

// New get a discovery.
func New(c *conf.Config) (d *Discovery, cancel context.CancelFunc) {
	d = &Discovery{
		protected: c.EnableProtect,
		c:         c,
		client:    http.NewClient(c.HTTPClient),
		registry:  registry.NewRegistry(c),
	}
	d.nodes.Store(registry.NewNodes(c))
	d.syncUp()
	cancel = d.regSelf()
	go d.nodesproc()
	go d.exitProtect()
	return
}

func (d *Discovery) exitProtect() {
	// exist protect mode after two renew cycle
	time.Sleep(time.Second * 60)
	d.Lock()
	d.protected = false
	d.Unlock()
}
