package http

import (
	"fmt"
	"net/http/httputil"
	"os"
	"runtime"
	"time"

	"github.com/Bilibili/discovery/conf"
	"github.com/Bilibili/discovery/discovery"
	"github.com/Bilibili/discovery/errors"
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
)

var (
	dis *discovery.Discovery
)

// Init init http
func Init(c *conf.Config) {
	initService(c)
	engine := gin.New()
	gin.SetMode(gin.ReleaseMode)
	engine.Use(loggerHandler, recoverHandler)
	innerRouter(engine)
	engine.Run(c.HTTPServer.Addr)
}

func initService(c *conf.Config) {
	dis = discovery.New(c)
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

func loggerHandler(c *gin.Context) {
	// Start timer
	start := time.Now()
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery
	method := c.Request.Method

	// Process request
	c.Next()

	// Stop timer
	end := time.Now()
	latency := end.Sub(start)
	statusCode := c.Writer.Status()
	comment := c.Errors.ByType(gin.ErrorTypePrivate).String()
	clientIP := c.ClientIP()
	if raw != "" {
		path = path + "?" + raw
	}
	log.Infof("METHOD:%s | PATH:%s | CODE:%d | IP:%s | TIME:%2f | ERROR:%s", method, path, statusCode, clientIP, latency, comment)
}

func recoverHandler(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			httprequest, _ := httputil.DumpRequest(c.Request, false)
			pnc := fmt.Sprintf("[Recovery] %s panic recovered:\n%s\n%s\n%s", time.Now().Format("2006-01-02 15:04:05"), string(httprequest), err, buf)
			fmt.Fprintf(os.Stderr, pnc)
			log.Error(pnc)
			c.AbortWithStatus(500)
		}
	}()
	c.Next()
}

type resp struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func result(c *gin.Context, data interface{}, err error) {
	c.JSON(200, resp{
		Code: errors.Code(err).Code(),
		Data: data,
	})
}
