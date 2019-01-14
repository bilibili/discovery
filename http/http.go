package http

import (
	"fmt"
	"net/http/httputil"
	"os"
	"runtime"
	"time"

	"discovery/conf"
	"discovery/discovery"
	"discovery/errors"

	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
)

const (
	contextErrCode = "context/err/code"
)

var (
	dis *discovery.Discovery
)

// Init init http
func Init(c *conf.Config, d *discovery.Discovery) {
	dis = d
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(loggerHandler, recoverHandler)
	innerRouter(engine)
	go func() {
		if err := engine.Run(c.HTTPServer.Addr); err != nil {
			panic(err)
		}
	}()
}

// innerRouter init local router api path.
func innerRouter(e *gin.Engine) {
	group := e.Group("/discovery")
	group.POST("/register", register)
	group.POST("/renew", renew)
	group.POST("/cancel", cancel)
	group.GET("/fetch/all", fetchAll)
	group.GET("/fetch", fetch)
	group.GET("/fetchs", fetchs)
	group.GET("/poll", poll)
	group.GET("/polls", polls)
	group.GET("/nodes", nodes)
	group.POST("/set", set)
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
	ecode := c.GetInt(contextErrCode)
	clientIP := c.ClientIP()
	if raw != "" {
		path = path + "?" + raw
	}
	log.Infof("METHOD:%s | PATH:%s | CODE:%d | IP:%s | TIME:%d | ECODE:%d", method, path, statusCode, clientIP, latency/time.Millisecond, ecode)
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
	ee := errors.Code(err)
	c.Set(contextErrCode, ee.Code())
	c.JSON(200, resp{
		Code: ee.Code(),
		Data: data,
	})
}
