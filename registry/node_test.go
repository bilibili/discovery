package registry

import (
	"context"
	"strings"
	"testing"

	dc "github.com/Bilibili/discovery/conf"
	"github.com/Bilibili/discovery/errors"
	"github.com/Bilibili/discovery/lib/http"
	"github.com/Bilibili/discovery/model"

	. "github.com/smartystreets/goconvey/convey"
	gock "gopkg.in/h2non/gock.v1"
)

func TestReplicate(t *testing.T) {
	Convey("test replicate", t, func() {
		i := model.NewInstance(reg)
		nodes := NewNodes(&dc.Config{HTTPClient: &http.ClientConfig{},
			HTTPServer: &dc.ServerConfig{Addr: "127.0.0.1:7171"},
			Nodes:      []string{"127.0.0.1:7172", "127.0.0.1:7173", "127.0.0.1:7171"},
		})
		nodes.nodes[0].client.SetTransport(gock.DefaultTransport)
		nodes.nodes[1].client.SetTransport(gock.DefaultTransport)
		httpMock("POST", "http://127.0.0.1:7172/discovery/register").Reply(200).JSON(`{"code":0}`)
		httpMock("POST", "http://127.0.0.1:7173/discovery/register").Reply(200).JSON(`{"code":0}`)
		err := nodes.Replicate(context.TODO(), model.Register, i, false)
		So(err, ShouldBeNil)
		err = nodes.Replicate(context.TODO(), model.Renew, i, false)
		So(err, ShouldBeNil)
		err = nodes.Replicate(context.TODO(), model.Cancel, i, false)
		So(err, ShouldBeNil)
	})
}

func TestNodes(t *testing.T) {
	Convey("test nodes", t, func() {
		nodes := NewNodes(&dc.Config{HTTPClient: &http.ClientConfig{},
			HTTPServer: &dc.ServerConfig{Addr: "127.0.0.1:7171"},
			Nodes:      []string{"127.0.0.1:7172", "127.0.0.1:7173", "127.0.0.1:7171"},
			Zones:      map[string]string{"zone": "127.0.0.1:7172"},
		})
		res := nodes.Nodes()
		So(len(res), ShouldResemble, 3)
	})
	Convey("test all nodes", t, func() {
		nodes := NewNodes(&dc.Config{HTTPClient: &http.ClientConfig{},
			HTTPServer: &dc.ServerConfig{Addr: "127.0.0.1:7171"},
			Nodes:      []string{"127.0.0.1:7172", "127.0.0.1:7173", "127.0.0.1:7171"},
			Zones:      map[string]string{"zone": "127.0.0.1:7172"},
		})
		res := nodes.AllNodes()
		So(len(res), ShouldResemble, 4)
	})
}

func TestUp(t *testing.T) {
	Convey("test up", t, func() {
		nodes := NewNodes(&dc.Config{HTTPClient: &http.ClientConfig{},
			HTTPServer: &dc.ServerConfig{Addr: "127.0.0.1:7171"},
			Nodes:      []string{"127.0.0.1:7172", "127.0.0.1:7173", "127.0.0.1:7171"},
		})
		nodes.UP()
		for _, nd := range nodes.nodes {
			if nd.addr == "127.0.0.1:7171" {
				So(nd.status, ShouldResemble, model.NodeStatusUP)
			}
		}
	})
}

func TestCall(t *testing.T) {
	Convey("test call", t, func() {
		var res *model.Instance
		node := newNode(&dc.Config{HTTPClient: &http.ClientConfig{},
			HTTPServer: &dc.ServerConfig{Addr: "127.0.0.1:7171"},
			Nodes:      []string{"127.0.0.1:7171"}}, "127.0.0.1:7172")
		node.client.SetTransport(gock.DefaultTransport)
		httpMock("POST", "http://127.0.0.1:7172/discovery/register").Reply(200).JSON(`{"ts":1514341945,"code":-409,"data":{"region":"shsb","zone":"fuck","appid":"main.arch.account-service","env":"pre","hostname":"cs4sq","http":"","rpc":"0.0.0.0:18888","weight":2}}`)
		i := model.NewInstance(reg)
		err := node.call(context.TODO(), model.Register, i, "http://127.0.0.1:7172/discovery/register", &res)
		So(err, ShouldResemble, errors.Conflict)
		So(res.AppID, ShouldResemble, "main.arch.account-service")
	})
}

func TestNodeCancel(t *testing.T) {
	Convey("test node renew 409 error", t, func() {
		i := model.NewInstance(reg)
		node := newNode(&dc.Config{HTTPClient: &http.ClientConfig{}, Nodes: []string{"127.0.0.1:7171", "127.0.0.1:7172"}}, "127.0.0.1:7172")
		node.pRegisterURL = "http://127.0.0.1:7171/discovery/register"
		node.client.SetTransport(gock.DefaultTransport)
		httpMock("POST", "http://127.0.0.1:7172/discovery/cancel").Reply(200).JSON(`{"code":0}`)
		err := node.Cancel(context.TODO(), i)
		So(err, ShouldBeNil)
	})
}

func TestNodeRenew(t *testing.T) {
	Convey("test node renew 409 error", t, func() {
		i := model.NewInstance(reg)
		node := newNode(&dc.Config{HTTPClient: &http.ClientConfig{}}, "127.0.0.1:7172")
		node.pRegisterURL = "http://127.0.0.1:7171/discovery/register"
		node.client.SetTransport(gock.DefaultTransport)
		httpMock("POST", "http://127.0.0.1:7172/discovery/renew").Reply(200).JSON(`{"code":-409,"data":{"region":"shsb","zone":"fuck","appid":"main.arch.account-service","env":"pre","hostname":"cs4sq","http":"","rpc":"0.0.0.0:18888","weight":2}}`)
		httpMock("POST", "http://127.0.0.1:7171/discovery/register").Reply(200).JSON(`{"code":0}`)
		err := node.Renew(context.TODO(), i)
		So(err, ShouldBeNil)
	})
}

func TestNodeRenew2(t *testing.T) {
	Convey("test node renew 404 error", t, func() {
		i := model.NewInstance(reg)
		node := newNode(&dc.Config{HTTPClient: &http.ClientConfig{}, HTTPServer: &dc.ServerConfig{Addr: "127.0.0.1:7171"}, Nodes: []string{"127.0.0.1:7171"}}, "127.0.0.1:7172")
		node.client.SetTransport(gock.DefaultTransport)
		httpMock("POST", "http://127.0.0.1:7172/discovery/renew").Reply(200).JSON(`{"code":-404}`)
		httpMock("POST", "http://127.0.0.1:7171/discovery/register").Reply(200).JSON(`{"code":0}`)
		err := node.Renew(context.TODO(), i)
		So(err, ShouldBeNil)
	})
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}
