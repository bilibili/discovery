package naming_test

import (
	"fmt"
	"time"

	"github.com/bilibili/discovery/naming"
)

// This Example register a server provider into discovery.
func ExampleDiscovery_Register() {
	conf := &naming.Config{
		Nodes: []string{"127.0.0.1:7171"}, // NOTE: 配置种子节点(1个或多个)，client内部可根据/discovery/nodes节点获取全部node(方便后面增减节点)
		Zone:  "sh1",
		Env:   "test",
	}
	dis := naming.New(conf)
	ins := &naming.Instance{
		Zone:  "sh1",
		Env:   "test",
		AppID: "provider",
		// Hostname:"", // NOTE: hostname 不需要，会优先使用discovery new时Config配置的值，如没有则从os.Hostname方法获取！！！
		Addrs:    []string{"http://172.0.0.1:8888", "grpc://172.0.0.1:9999"},
		LastTs:   time.Now().Unix(),
		Metadata: map[string]string{"weight": "10"},
	}
	cancel, _ := dis.Register(ins)
	defer cancel() // NOTE: 注意一般在进程退出的时候执行，会调用discovery的cancel接口，使实例从discovery移除
	fmt.Println("register")
	// Unordered output4
}

type consumer struct {
	conf  *naming.Config
	appID string
	dis   naming.Resolver
	ins   []*naming.Instance
}

// This Example show how get watch a server provider and get provider instances.
func ExampleResolver_Watch() {
	conf := &naming.Config{
		Nodes: []string{"127.0.0.1:7171"},
		Zone:  "sh1",
		Env:   "test",
	}
	dis := naming.New(conf)
	c := &consumer{
		conf:  conf,
		appID: "provider",
		dis:   dis.Build("provider"),
	}
	rsl := dis.Build(c.appID)
	ch := rsl.Watch()
	go c.getInstances(ch)
	in := c.getInstance()
	_ = in
}

func (c *consumer) getInstances(ch <-chan struct{}) {
	for { // NOTE: 通过watch返回的event chan =>
		if _, ok := <-ch; !ok {
			return
		}
		// NOTE: <= 实时fetch最新的instance实例
		ins, ok := c.dis.Fetch()
		if !ok {
			continue
		}
		// get local zone instances, otherwise get all zone instances.
		if in, ok := ins.Instances[c.conf.Zone]; ok {
			c.ins = in
		} else {
			for _, in := range ins.Instances {
				c.ins = append(c.ins, in...)
			}
		}
	}
}

func (c *consumer) getInstance() (ins *naming.Instance) {
	// get instance by loadbalance
	// you can use any loadbalance algorithm what you want.
	return
}
