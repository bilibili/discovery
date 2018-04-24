package http

import (
	"github.com/Bilibili/discovery/conf"
	"github.com/Bilibili/discovery/service"
	gin "github.com/gin-gonic/gin"
)

var (
	svr *service.Service
)

// Init init http
func Init(c *conf.Config) {
	initService(c)
	engine := gin.Default()
	innerRouter(engine)
	engine.Run(":8080")
}

func initService(c *conf.Config) {
	svr = service.New(c)
}

// innerRouter init local router api path.
func innerRouter(e *gin.Engine) {
	group := e.Group("/discovery")
	group.POST("/register", register)
	group.POST("/renew", renew)
	group.POST("/cancel", cancel)
	group.GET("/fetch/all", fetchAll)
	group.GET("/fetch", fetch)
	group.GET("/poll", poll)
	group.GET("/polls", polls)
	group.GET("/nodes", nodes)
}
