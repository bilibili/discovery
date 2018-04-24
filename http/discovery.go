package http

import (
	"encoding/json"
	"net/http"
	"time"

	log "golang/log4go"

	"github.com/Bilibili/discovery/errors"
	"github.com/Bilibili/discovery/model"
	gin "github.com/gin-gonic/gin"
)

const (
	_pollWaitSecond = 30 * time.Second
)

type resp struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func register(c *gin.Context) {
	arg := new(model.ArgRegister)
	if err := c.Bind(arg); err != nil {
		return
	}
	i := model.NewInstance(arg)
	if i.Status == 0 || i.Status > 2 {
		c.JSON(200, resp{
			Code: -400,
		})
		return
	}
	if arg.Metadata != "" {
		i.Metadata = json.RawMessage([]byte(arg.Metadata))
		// check the metadata type is json
		if !json.Valid([]byte(arg.Metadata)) {
			c.JSON(200, resp{
				Code: -400,
			})
			log.Error("register params() metadata(%v) invalid json", arg.Metadata)
			return
		}
	}
	// register replication
	if arg.DirtyTimestamp > 0 {
		i.DirtyTimestamp = arg.DirtyTimestamp
	}
	svr.Register(c, i, arg)
	c.JSON(200, resp{})
}

func renew(c *gin.Context) {
	arg := new(model.ArgRenew)
	if err := c.Bind(arg); err != nil {
		return
	}
	// renew
	instance, err := svr.Renew(c, arg)
	if err != nil {
		c.JSON(200, resp{
			Code: errors.Code(err),
		})
		return
	}
	c.JSON(200, resp{
		Data: instance,
	})
}

func cancel(c *gin.Context) {
	arg := new(model.ArgCancel)
	if err := c.Bind(arg); err != nil {
		c.JSON(200, resp{
			Code: -400,
		})
		return
	}
	// cancel
	if err := svr.Cancel(c, arg); err != nil {
		c.JSON(200, resp{
			Code: errors.Code(err),
		})
		return
	}
	c.JSON(200, resp{})
}

func fetchAll(c *gin.Context) {
	c.JSON(200, resp{Data: svr.FetchAll(c)})
}

func fetch(c *gin.Context) {
	arg := new(model.ArgFetch)
	if err := c.Bind(arg); err != nil {
		return
	}
	insInfo, err := svr.Fetch(c, arg)
	if err != nil {
		c.JSON(200, resp{
			Code: errors.Code(err),
		})
		return
	}
	c.JSON(200, resp{Data: insInfo})
}

func poll(c *gin.Context) {
	arg := new(model.ArgPolls)
	if err := c.Bind(arg); err != nil {
		return
	}
	if len(arg.Treeid) != 1 {
		c.JSON(200, resp{
			Code: -400,
		})
		return
	}
	ch, _, err := svr.Polls(c, arg)
	if err != nil && err != errors.NotModified {
		c.JSON(200, resp{
			Code: errors.Code(err),
		})
		return
	}
	// wait for instance change
	select {
	case e := <-ch:
		c.JSON(200, resp{Data: e[arg.Treeid[0]]})
		svr.PutChan(ch)
		// broadcast will delete all connections of appid
	case <-time.After(_pollWaitSecond):
		c.JSON(200, resp{
			Code: -304,
		})
		svr.DelConns(arg)
	case <-c.Writer.(http.CloseNotifier).CloseNotify():
		c.JSON(200, resp{
			Code: -304,
		})
		svr.DelConns(arg)
	}
}

func polls(c *gin.Context) {
	arg := new(model.ArgPolls)
	if err := c.Bind(arg); err != nil {
		return
	}
	if len(arg.Treeid) != len(arg.LatestTimestamp) && len(arg.Appid) != len(arg.LatestTimestamp) {
		c.JSON(200, resp{
			Code: -400,
		})
		return
	}
	ch, new, err := svr.Polls(c, arg)
	if err != nil && err != errors.NotModified {
		c.JSON(200, resp{
			Code: errors.Code(err),
		})
		return
	}
	// wait for instance change
	select {
	case e := <-ch:
		c.JSON(200, resp{
			Data: e,
		})
		if new {
			svr.PutChan(ch)
		} else {
			svr.DelConns(arg)
		}
		// broadcast will delete all connections of appid
	case <-time.After(_pollWaitSecond):
		c.JSON(200, resp{
			Code: -304,
		})
		svr.DelConns(arg)
	case <-c.Writer.(http.CloseNotifier).CloseNotify():
		c.JSON(200, resp{
			Code: -304,
		})
		svr.DelConns(arg)
	}
}

func nodes(c *gin.Context) {
	c.JSON(200, resp{Data: svr.Nodes(c)})
}
