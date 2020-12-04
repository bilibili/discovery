package http

import (
	"errors"

	"github.com/bilibili/discovery/conf"
	"github.com/bilibili/discovery/discovery"

	log "github.com/go-kratos/kratos/pkg/log"
	bm "github.com/go-kratos/kratos/pkg/net/http/blademaster"
)

var (
	dis          *discovery.Discovery
	protected    = true
	errProtected = errors.New("discovery in protect mode and only support register")
)

// Init init http
func Init(c *conf.Config, s *discovery.Discovery) {
	dis = s
	engineInner := bm.DefaultServer(c.HTTPServer)
	innerRouter(engineInner)
	if err := engineInner.Start(); err != nil {
		log.Error("bm.DefaultServer error(%v)", err)
		panic(err)
	}
	log.Info("[HTTP] Listening on: %s", c.HTTPServer.Addr)
}

// innerRouter init local router api path.
func innerRouter(e *bm.Engine) {
	group := e.Group("/discovery")
	{
		group.POST("/register", register)
		group.POST("/renew", renew)
		group.POST("/cancel", cancel)
		group.GET("/fetch/all", initProtect, fetchAll)
		group.GET("/fetch", initProtect, fetch)
		group.GET("/fetchs", initProtect, fetchs)
		group.GET("/poll", initProtect, poll)
		group.GET("/polls", initProtect, polls)
		//manager
		group.POST("/set", set)
		group.GET("/nodes", initProtect, nodes)
	}
}

func initProtect(ctx *bm.Context) {
	if dis.Protected() {
		ctx.JSON(nil, errProtected)
		ctx.AbortWithStatus(503)
	}
}
