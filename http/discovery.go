package http

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bilibili/discovery/model"

	"github.com/go-kratos/kratos/pkg/ecode"
	log "github.com/go-kratos/kratos/pkg/log"
	bm "github.com/go-kratos/kratos/pkg/net/http/blademaster"
)

const (
	_pollWaitSecond = 30 * time.Second
)

func register(c *bm.Context) {
	arg := new(model.ArgRegister)
	if err := c.Bind(arg); err != nil {
		return
	}
	i := model.NewInstance(arg)
	if i.Status == 0 || i.Status > 2 {
		log.Error("register params status invalid")
		return
	}
	if arg.Metadata != "" {
		// check the metadata type is json
		if !json.Valid([]byte(arg.Metadata)) {
			c.JSON(nil, ecode.RequestErr)
			log.Error("register params() metadata(%v) invalid json", arg.Metadata)
			return
		}
	}
	// register replication
	if arg.DirtyTimestamp > 0 {
		i.DirtyTimestamp = arg.DirtyTimestamp
	}
	dis.Register(c, i, arg.LatestTimestamp, arg.Replication, arg.FromZone)
	c.JSON(nil, nil)
}

func renew(c *bm.Context) {
	arg := new(model.ArgRenew)
	if err := c.Bind(arg); err != nil {
		return
	}
	// renew
	c.JSON(dis.Renew(c, arg))
}

func cancel(c *bm.Context) {
	arg := new(model.ArgCancel)
	if err := c.Bind(arg); err != nil {
		return
	}
	// cancel
	c.JSON(nil, dis.Cancel(c, arg))
}

func fetchAll(c *bm.Context) {
	c.JSON(dis.FetchAll(c), nil)
}

func fetch(c *bm.Context) {
	arg := new(model.ArgFetch)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(dis.Fetch(c, arg))
}

func fetchs(c *bm.Context) {
	arg := new(model.ArgFetchs)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(dis.Fetchs(c, arg))
}

func poll(c *bm.Context) {
	arg := new(model.ArgPolls)
	if err := c.Bind(arg); err != nil {
		return
	}
	ch, new, miss, err := dis.Polls(c, arg)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	for _, mi := range miss {
		if mi == arg.AppID[0] {
			c.JSONMap(map[string]interface{}{
				"message": fmt.Sprintf("%s not found", mi),
			}, ecode.NothingFound)
			return
		}
	}
	// wait for instance change
	select {
	case e := <-ch:
		c.JSON(e[arg.AppID[0]], nil)
		if !new {
			dis.DelConns(arg) // broadcast will delete all connections of appid
		}
		return
	case <-time.After(_pollWaitSecond):
	case <-c.Done():
	}
	c.JSON(nil, ecode.NotModified)
	dis.DelConns(arg)
}

func polls(c *bm.Context) {
	arg := new(model.ArgPolls)
	if err := c.Bind(arg); err != nil {
		return
	}
	if len(arg.AppID) != len(arg.LatestTimestamp) {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	ch, new, miss, err := dis.Polls(c, arg)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	// wait for instance change
	select {
	case e := <-ch:
		c.JSONMap(map[string]interface{}{
			"data": e,
			"error": map[ecode.Code]interface{}{
				ecode.NothingFound: miss,
			},
		}, nil)
		if !new {
			dis.DelConns(arg) // broadcast will delete all connections of appid
		}
		return
	case <-time.After(_pollWaitSecond):
	case <-c.Done():
	}
	c.JSON(nil, ecode.NotModified)
	dis.DelConns(arg)
}

func set(c *bm.Context) {
	arg := new(model.ArgSet)
	if err := c.Bind(arg); err != nil {
		return
	}
	// len of color,status,metadata must equal to len of hostname or be zero
	if (len(arg.Hostname) != len(arg.Status) && len(arg.Status) != 0) ||
		(len(arg.Hostname) != len(arg.Metadata) && len(arg.Metadata) != 0) {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if arg.SetTimestamp == 0 {
		arg.SetTimestamp = time.Now().UnixNano()
	}
	c.JSON(nil, dis.Set(c, arg))
}

func nodes(c *bm.Context) {
	c.JSON(dis.Nodes(c), nil)
}
