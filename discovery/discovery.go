package discovery

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/bilibili/discovery/conf"
	"github.com/bilibili/discovery/registry"
	http "github.com/go-kratos/kratos/pkg/net/http/blademaster"
)

// Discovery discovery.
type Discovery struct {
	c         *conf.Config
	protected bool
	client    *http.Client
	registry  *registry.Registry
	nodes     atomic.Value
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
	d.protected = false
}
