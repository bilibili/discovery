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
		result(c, nil, errors.ParamsErr)
		log.Error("register params status invalid")
		return
	}
	if arg.Metadata != "" {
		// check the metadata type is json
		if !json.Valid([]byte(arg.Metadata)) {
			result(c, nil, errors.ParamsErr)
			log.Error("register params() metadata(%v) invalid json", arg.Metadata)
			return
		}
	}
	// register replication
	if arg.DirtyTimestamp > 0 {
		i.DirtyTimestamp = arg.DirtyTimestamp
	}
	svr.Register(c, i, arg)
	result(c, nil, nil)
}

func renew(c *gin.Context) {
	arg := new(model.ArgRenew)
	if err := c.Bind(arg); err != nil {
		return
	}
	// renew
	instance, err := svr.Renew(c, arg)
	result(c, instance, err)
}

func cancel(c *gin.Context) {
	arg := new(model.ArgCancel)
	if err := c.Bind(arg); err != nil {
		result(c, nil, errors.ParamsErr)
		return
	}
	// cancel
	result(c, nil, svr.Cancel(c, arg))
}

func result(c *gin.Context, data interface{}, err error) {
	c.JSON(200, resp{
		Code: errors.Code(err),
		Data: data,
	})
}

func fetchAll(c *gin.Context) {
	result(c, svr.FetchAll(c), nil)
}

func fetch(c *gin.Context) {
	arg := new(model.ArgFetch)
	if err := c.Bind(arg); err != nil {
		result(c, nil, errors.ParamsErr)
		return
	}
	insInfo, err := svr.Fetch(c, arg)
	result(c, insInfo, err)
}

func poll(c *gin.Context) {
	arg := new(model.ArgPolls)
	if err := c.Bind(arg); err != nil {
		return
	}
	ch, _, err := svr.Polls(c, arg)
	if err != nil && err != errors.NotModified {
		result(c, nil, err)
		return
	}
	// wait for instance change
	select {
	case e := <-ch:
		result(c, resp{Data: e[arg.Appid[0]]}, nil)
		svr.PutChan(ch)
		// broadcast will delete all connections of appid
	case <-time.After(_pollWaitSecond):
		result(c, nil, errors.NotModified)
		svr.DelConns(arg)
	case <-c.Writer.(http.CloseNotifier).CloseNotify():
		result(c, nil, errors.NotModified)
		svr.DelConns(arg)
	}
}

func polls(c *gin.Context) {
	arg := new(model.ArgPolls)
	if err := c.Bind(arg); err != nil {
		return
	}
	if len(arg.Appid) != len(arg.LatestTimestamp) {
		result(c, nil, errors.ParamsErr)
		return
	}
	ch, new, err := svr.Polls(c, arg)
	if err != nil && err != errors.NotModified {
		result(c, nil, err)
		return
	}
	// wait for instance change
	select {
	case e := <-ch:
		result(c, e, nil)
		if new {
			svr.PutChan(ch)
		} else {
			svr.DelConns(arg)
		}
		// broadcast will delete all connections of appid
	case <-time.After(_pollWaitSecond):
		result(c, nil, errors.NotModified)
		svr.DelConns(arg)
	case <-c.Writer.(http.CloseNotifier).CloseNotify():
		result(c, nil, errors.NotModified)
		svr.DelConns(arg)
	}
}

func nodes(c *gin.Context) {
	c.JSON(200, resp{Data: svr.Nodes(c)})
}
