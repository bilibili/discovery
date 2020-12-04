package discovery

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"
	"time"

	dc "github.com/bilibili/discovery/conf"
	"github.com/bilibili/discovery/model"

	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/ecode"
	http "github.com/go-kratos/kratos/pkg/net/http/blademaster"
	xtime "github.com/go-kratos/kratos/pkg/time"
	. "github.com/smartystreets/goconvey/convey"
	gock "gopkg.in/h2non/gock.v1"
)

var (
	ctx    = context.TODO()
	reg    = defRegisArg()
	rew    = &model.ArgRenew{AppID: "main.arch.test", Hostname: "test1", Zone: "sh001", Env: "pre"}
	rew2   = &model.ArgRenew{AppID: "main.arch.test", Hostname: "test1", Zone: "sh001", Env: "pre"}
	cancel = &model.ArgCancel{AppID: "main.arch.test", Hostname: "test1", Zone: "sh001", Env: "pre"}
	fet    = &model.ArgFetch{AppID: "main.arch.test", Zone: "sh001", Env: "pre", Status: 1}
	set    = &model.ArgSet{AppID: "main.arch.test",
		Zone: "sh001", Env: "pre",
		Hostname: []string{"test1"},
		Status:   []int64{1},
	}
	pollArg = newPoll()
)

func TestMain(m *testing.M) {
	flag.Set("conf", "./")
	flag.Parse()
	paladin.Init()
	m.Run()
	os.Exit(0)
}
func newFetchArg() *model.ArgFetchs {
	return &model.ArgFetchs{AppID: []string{"main.arch.test"}, Zone: "sh001", Env: "pre", Status: 1}
}
func newPoll() *model.ArgPolls {
	return &model.ArgPolls{
		Env:             "pre",
		AppID:           []string{"main.arch.test"},
		LatestTimestamp: []int64{0},
	}
}
func defRegisArg() *model.ArgRegister {
	return &model.ArgRegister{
		AppID:           "main.arch.test",
		Hostname:        "test1",
		Zone:            "sh001",
		Env:             "pre",
		Status:          1,
		Metadata:        `{"test":"test","weight":"10"}`,
		LatestTimestamp: time.Now().UnixNano(),
	}
}
func defRegDiscovery() *model.Instance {
	return &model.Instance{
		AppID:           "infra.discovery",
		Hostname:        "test2",
		Zone:            "sh001",
		Env:             "pre",
		Status:          1,
		Addrs:           []string{"http://127.0.0.1:7172"},
		LatestTimestamp: time.Now().UnixNano(),
	}
}

var config = newConfig()

func newConfig() *dc.Config {
	c := &dc.Config{
		HTTPClient: &http.ClientConfig{
			Timeout:   xtime.Duration(time.Second * 30),
			Dial:      xtime.Duration(time.Second),
			KeepAlive: xtime.Duration(time.Second * 30),
		},
		HTTPServer: &http.ServerConfig{Addr: "127.0.0.1:7171"},
		Nodes:      []string{"127.0.0.1:7171", "127.0.0.1:7172"},
		Env: &dc.Env{
			Zone:      "sh001",
			DeployEnv: "pre",
			Host:      "test_server",
		},
	}
	return c
}
func init() {
	httpMock("GET", "http://127.0.0.1:7172/discovery/fetch/all").Reply(200).JSON(`{"code":0}`)
	httpMock("POST", "http://127.0.0.1:7172/discovery/register").Reply(200).JSON(`{"code":0}`)
	httpMock("POST", "http://127.0.0.1:7172/discovery/cancel").Reply(200).JSON(`{"code":0}`)

	os.Setenv("ZONE", "sh001")
	os.Setenv("DEPLOY_ENV", "pre")
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestRegister(t *testing.T) {
	Convey("test Register", t, func() {
		svr, cancel := New(config)
		defer cancel()
		svr.client.SetTransport(gock.DefaultTransport)
		svr.syncUp()
		i := model.NewInstance(reg)
		svr.Register(context.TODO(), i, reg.LatestTimestamp, reg.Replication, true)
		ins, err := svr.Fetch(context.TODO(), fet)
		So(err, ShouldBeNil)
		So(len(ins.Instances), ShouldResemble, 1)
		Convey("test metadta", func() {
			for _, is := range ins.Instances {
				So(err, ShouldBeNil)
				for _, i := range is {
					So(i.Metadata["weight"], ShouldEqual, "10")
					So(i.Metadata["test"], ShouldEqual, "test")
				}
			}
		})
		Convey("test set", func() {
			err = svr.Set(context.TODO(), set)
			So(err, ShouldBeNil)
			ins, err = svr.Fetch(context.TODO(), fet)
			So(err, ShouldBeNil)
			So(len(ins.Instances), ShouldResemble, 1)
			for _, is := range ins.Instances {
				for _, i := range is {
					So(i.Status, ShouldEqual, 1)
				}
			}
		})
	})
}
func TestDiscovery(t *testing.T) {
	Convey("test cancel polls", t, func() {
		svr, disCancel := New(config)
		defer disCancel()
		svr.client.SetTransport(gock.DefaultTransport)
		reg2 := defRegisArg()
		reg2.Hostname = "test2"
		i1 := model.NewInstance(reg)
		i2 := model.NewInstance(reg2)
		svr.Register(context.TODO(), i1, reg.LatestTimestamp, reg.Replication, reg.FromZone)
		svr.Register(context.TODO(), i2, reg2.LatestTimestamp, reg.Replication, reg.FromZone)
		ch, new, _, err := svr.Polls(context.TODO(), pollArg)
		So(err, ShouldBeNil)
		So(new, ShouldBeTrue)
		ins := <-ch
		So(len(ins["main.arch.test"].Instances["sh001"]), ShouldEqual, 2)
		pollArg.LatestTimestamp[0] = ins["main.arch.test"].LatestTimestamp
		time.Sleep(time.Second)
		err = svr.Cancel(context.TODO(), cancel)
		So(err, ShouldBeNil)
		ch, new, _, err = svr.Polls(context.TODO(), pollArg)
		So(err, ShouldBeNil)
		So(new, ShouldBeTrue)
		ins = <-ch
		So(len(ins["main.arch.test"].Instances), ShouldEqual, 1)
	})
}
func TestFetchs(t *testing.T) {
	Convey("test fetch multi appid", t, func() {
		svr, cancel := New(config)
		defer cancel()
		svr.client.SetTransport(gock.DefaultTransport)
		reg2 := defRegisArg()
		reg2.AppID = "appid2"
		i1 := model.NewInstance(reg)
		i2 := model.NewInstance(reg2)
		svr.Register(context.TODO(), i1, reg.LatestTimestamp, reg.Replication, reg.FromZone)
		svr.Register(context.TODO(), i2, reg2.LatestTimestamp, reg.Replication, reg.FromZone)
		fetchs := newFetchArg()
		fetchs.AppID = append(fetchs.AppID, "appid2")
		is, err := svr.Fetchs(ctx, fetchs)
		So(err, ShouldBeNil)
		So(len(is), ShouldResemble, 2)
	})
}
func TestZones(t *testing.T) {
	Convey("test multi zone discovery", t, func() {
		svr, cancel := New(config)
		defer cancel()
		svr.client.SetTransport(gock.DefaultTransport)
		reg2 := defRegisArg()
		reg2.Zone = "sh002"
		i1 := model.NewInstance(reg)
		i2 := model.NewInstance(reg2)
		svr.Register(context.TODO(), i1, reg.LatestTimestamp, reg.Replication, reg.FromZone)
		svr.Register(context.TODO(), i2, reg2.LatestTimestamp, reg2.Replication, reg2.FromZone)
		ch, new, _, err := svr.Polls(context.TODO(), newPoll())
		So(err, ShouldBeNil)
		So(new, ShouldBeTrue)
		ins := <-ch
		So(len(ins["main.arch.test"].Instances), ShouldEqual, 2)
		pollArg.Zone = "sh002"
		ch, new, _, err = svr.Polls(context.TODO(), newPoll())
		So(err, ShouldBeNil)
		So(new, ShouldBeTrue)
		ins = <-ch
		So(len(ins["main.arch.test"].Instances["sh002"]), ShouldEqual, 1)
		Convey("test zone update", func() {
			pollArg.LatestTimestamp = []int64{ins["main.arch.test"].LatestTimestamp}
			pollArg.Zone = ""
			reg3 := defRegisArg()
			reg3.Zone = "sh002"
			reg3.Hostname = "test03"
			i3 := model.NewInstance(reg3)
			svr.Register(context.TODO(), i3, reg3.LatestTimestamp, reg3.Replication, reg3.FromZone)
			ch, _, _, err = svr.Polls(context.TODO(), pollArg)
			So(err, ShouldBeNil)
			ins = <-ch
			So(len(ins["main.arch.test"].Instances), ShouldResemble, 2)
			So(len(ins["main.arch.test"].Instances["sh002"]), ShouldResemble, 2)
			So(len(ins["main.arch.test"].Instances["sh001"]), ShouldResemble, 1)
			pollArg.LatestTimestamp = []int64{ins["main.arch.test"].LatestTimestamp}
			_, _, _, err = svr.Polls(context.TODO(), pollArg)
			So(err, ShouldResemble, ecode.NotModified)
		})
	})
}
func TestRenew(t *testing.T) {
	Convey("test Renew", t, func() {
		svr, cancel := New(config)
		defer cancel()
		svr.client.SetTransport(gock.DefaultTransport)
		i := model.NewInstance(reg)
		svr.Register(context.TODO(), i, reg.LatestTimestamp, reg.Replication, reg.FromZone)
		_, err := svr.Renew(context.TODO(), rew)
		So(err, ShouldBeNil)
		rew2.AppID = "main.arch.noexist"
		_, err = svr.Renew(context.TODO(), rew2)
		So(err, ShouldResemble, ecode.NothingFound)
		rew2.AppID = "main.arch.test"
		rew2.DirtyTimestamp = 1
		rew2.Replication = true
		_, err = svr.Renew(context.TODO(), rew2)
		So(err, ShouldResemble, ecode.Conflict)
		rew2.DirtyTimestamp = time.Now().UnixNano()
		_, err = svr.Renew(context.TODO(), rew2)
		So(err, ShouldResemble, ecode.NothingFound)
	})
}

func TestCancel(t *testing.T) {
	Convey("test cancel", t, func() {
		svr, disCancel := New(config)
		defer disCancel()
		svr.client.SetTransport(gock.DefaultTransport)
		i := model.NewInstance(reg)
		svr.Register(context.TODO(), i, reg.LatestTimestamp, reg.Replication, reg.FromZone)
		err := svr.Cancel(context.TODO(), cancel)
		So(err, ShouldBeNil)
		err = svr.Cancel(context.TODO(), cancel)
		So(err, ShouldResemble, ecode.NothingFound)
		_, err = svr.Fetch(context.TODO(), fet)
		So(err, ShouldResemble, ecode.NothingFound)
	})
}

func TestFetchAll(t *testing.T) {
	Convey("test fetch all", t, func() {
		svr, cancel := New(config)
		defer cancel()
		svr.client.SetTransport(gock.DefaultTransport)
		i := model.NewInstance(reg)
		svr.Register(context.TODO(), i, reg.LatestTimestamp, reg.Replication, reg.FromZone)
		fs := svr.FetchAll(context.TODO())[i.AppID]
		So(len(fs), ShouldResemble, 1)
	})
}

func TestNodes(t *testing.T) {
	Convey("test nodes", t, func() {
		svr, cancel := New(config)
		defer cancel()
		svr.client.SetTransport(gock.DefaultTransport)
		svr.Register(context.Background(), defRegDiscovery(), time.Now().UnixNano(), false, true)
		time.Sleep(time.Second)
		ns := svr.Nodes(context.TODO())
		So(len(ns), ShouldResemble, 2)
	})
}
