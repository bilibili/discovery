package naming

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Bilibili/discovery/conf"
	"github.com/Bilibili/discovery/http"
	xhttp "github.com/Bilibili/discovery/lib/http"
	xtime "github.com/Bilibili/discovery/lib/time"

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
	os.Setenv("ZONE", "test")
	os.Setenv("DEPLOY_ENV", "test")
	conf := &Config{
		Nodes: []string{"127.0.0.1:7171"},
	}
	dis := New(conf)
	appid := "test1"
	Convey("test discovery register", t, func() {
		instance := &Instance{
			Zone:     "test",
			Env:      "test",
			AppID:    appid,
			Hostname: "test-host",
		}
		_, err := dis.Register(instance)
		So(err, ShouldBeNil)
		conf.Nodes = []string{"127.0.0.1:7172"}
		dis.Reload(conf)
		instance.AppID = "test2"
		//instance.Metadata = map[string]string{"meta": "meta"}
		_, err = dis.Register(instance)
		So(err, ShouldNotBeNil)
		dis.renew(context.TODO(), instance)
		conf.Nodes = []string{"127.0.0.1:7171"}
		dis.Reload(conf)
		dis.renew(context.TODO(), instance)
	})
	Convey("test discovery watch", t, func() {
		ch := dis.Watch(appid)
		<-ch
		ins, ok := dis.Fetch(appid)
		So(ok, ShouldBeTrue)
		So(len(ins["test"]), ShouldEqual, 1)
		So(ins["test"][0].AppID, ShouldEqual, appid)
		instance2 := &Instance{
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
		dis.Unwatch(appid)
		_, ok = dis.Fetch(appid)
		So(ok, ShouldBeFalse)
		conf.Nodes = []string{"127.0.0.1:7172"}
		dis.Reload(conf)
		So(dis.Scheme(), ShouldEqual, "discovery")
		dis.Close()
	})
}

func addNewInstance(ins *Instance) error {
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
