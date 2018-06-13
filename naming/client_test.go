package naming_test

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Bilibili/discovery/conf"
	"github.com/Bilibili/discovery/http"
	xhttp "github.com/Bilibili/discovery/lib/http"
	xtime "github.com/Bilibili/discovery/lib/time"
	"github.com/Bilibili/discovery/naming"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMain(m *testing.M) {
	go mockDiscoverySvr()
	m.Run()
}
func mockDiscoverySvr() {
	http.Init(&conf.Config{
		Zone:  "test",
		Nodes: []string{"127.0.0.1:7171"},
		HTTPServer: &conf.ServerConfig{
			Addr: "127.0.0.1:7171",
		},
		HTTPClient: &xhttp.ClientConfig{
			Dial: xtime.Duration(1),
		},
	})
}
func TestDiscovery(t *testing.T) {
	conf := &naming.Config{
		Nodes: []string{"127.0.0.1:7171"},
		Zone:  "test",
		Env:   "test",
		Host:  "aa",
	}
	dis := naming.New(conf)
	appid := "test1"
	Convey("test discovery register", t, func() {
		instance := &naming.Instance{
			Zone:     "test",
			Env:      "test",
			AppID:    appid,
			Hostname: "test-host",
		}
		_, err := dis.Register(instance)
		So(err, ShouldBeNil)
	})
	Convey("test discovery watch", t, func() {
		ch := dis.Watch(appid)
		<-ch
		ins, ok := dis.Fetch(appid)
		So(ok, ShouldBeTrue)
		So(len(ins["test"]), ShouldEqual, 1)
		So(ins["test"][0].AppID, ShouldEqual, appid)
		instance2 := &naming.Instance{
			Zone:     "test",
			Env:      "test",
			AppID:    appid,
			Hostname: "test-host2",
		}
		err := addNewInstance(instance2)
		So(err, ShouldBeNil)
		// watch for next update
		<-ch
		ins, ok = dis.Fetch(appid)
		So(ok, ShouldBeTrue)
		So(len(ins["test"]), ShouldEqual, 2)
		So(ins["test"][0].AppID, ShouldEqual, appid)
	})
}

func addNewInstance(ins *naming.Instance) error {
	cli := xhttp.NewClient(&xhttp.ClientConfig{
		Dial:      xtime.Duration(time.Second),
		KeepAlive: xtime.Duration(time.Second * 30),
	})
	params := url.Values{}
	params.Set("env", ins.Env)
	params.Set("zone", ins.Zone)
	params.Set("hostname", ins.Hostname)
	params.Set("appid", ins.AppID)
	params.Set("addrs", strings.Join(ins.Addrs, ","))
	params.Set("color", ins.Color)
	params.Set("version", ins.Version)
	params.Set("status", "1")
	res := new(struct {
		Code int `json:"code"`
	})
	return cli.Post(context.TODO(), "http://127.0.0.1:7171/discovery/register", "", params, &res)
}
