package naming

import (
	"context"
	"flag"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Bilibili/discovery/conf"
	"github.com/Bilibili/discovery/discovery"
	"github.com/Bilibili/discovery/http"
	xhttp "github.com/Bilibili/discovery/lib/http"
	xtime "github.com/Bilibili/discovery/lib/time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMain(m *testing.M) {
	os.Setenv("ZONE", "test")
	os.Setenv("DEPLOY_ENV", "test")
	flag.Parse()
	go mockDiscoverySvr()
	time.Sleep(time.Second)
	m.Run()
}
func mockDiscoverySvr() {
	c := &conf.Config{
		Zone:  "test",
		Nodes: []string{"127.0.0.1:7171"},
		HTTPServer: &conf.ServerConfig{
			Addr: "127.0.0.1:7171",
		},
		HTTPClient: &xhttp.ClientConfig{
			Dial: xtime.Duration(1),
		},
	}
	c.Fix()
	dis, _ := discovery.New(c)
	http.Init(c, dis)
}
func TestDiscovery(t *testing.T) {
	conf := &Config{
		Nodes: []string{"127.0.0.1:7171"},
		Host:  "test",
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
		dis.node.Store([]string{"127.0.0.1:7172"})
		instance.AppID = "test2"
		//instance.Metadata = map[string]string{"meta": "meta"}
		_, err = dis.Register(instance)
		So(err, ShouldNotBeNil)
		dis.renew(context.TODO(), instance)
		dis.node.Store([]string{"127.0.0.1:7171"})
		dis.renew(context.TODO(), instance)
		Convey("test discovery set", func() {
			rs := dis.Build(appid)
			inSet := &Instance{
				Zone:     "test",
				Env:      "test",
				AppID:    appid,
				Hostname: "test-host",
				Addrs: []string{
					"grpc://127.0.0.1:8080",
				},
				Metadata: map[string]string{
					"test":   "1",
					"weight": "111",
					"color":  "blue",
				},
			}
			err = dis.Set(inSet)
			So(err, ShouldBeNil)
			ch := rs.Watch()
			<-ch
			ins, _ := rs.Fetch()
			So(ins["test"][0].Metadata["weight"], ShouldResemble, "111")
		})
	})
	Convey("test discovery watch", t, func() {
		rsl := dis.Build(appid)
		ch := rsl.Watch()
		<-ch
		ins, ok := rsl.Fetch()
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
		ins, ok = rsl.Fetch()
		So(ok, ShouldBeTrue)
		So(len(ins["test"]), ShouldEqual, 2)
		So(ins["test"][0].AppID, ShouldEqual, appid)
		rsl.Close()
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
	params.Set("version", ins.Version)
	params.Set("status", "1")
	params.Set("latest_timestamp", strconv.FormatInt(time.Now().UnixNano(), 10))
	res := new(struct {
		Code int `json:"code"`
	})
	return cli.Post(context.TODO(), "http://127.0.0.1:7171/discovery/register", "", params, &res)
}
