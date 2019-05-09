package resolver

import (
	"sync"
	"testing"
	"time"

	"github.com/bilibili/discovery/naming"

	"google.golang.org/grpc/resolver"
)

type mockResolver struct {
}

func (m *mockResolver) Fetch() (insInfo *naming.InstancesInfo, ok bool) {
	ins := make(map[string][]*naming.Instance)
	ins["sh001"] = []*naming.Instance{
		&naming.Instance{Addrs: []string{"http://127.0.0.1:8080", "grpc://127.0.0.1:9090"}, Metadata: map[string]string{naming.MetaCluster: "c1"}},
		&naming.Instance{Addrs: []string{"http://127.0.0.2:8080", "grpc://127.0.0.2:9090"}, Metadata: map[string]string{naming.MetaCluster: "c1"}},
		&naming.Instance{Addrs: []string{"http://127.0.0.3:8080", "grpc://127.0.0.3:9090"}, Metadata: map[string]string{naming.MetaCluster: "c1"}},
		&naming.Instance{Addrs: []string{"http://127.0.0.4:8080", "grpc://127.0.0.4:9090"}, Metadata: map[string]string{naming.MetaCluster: "c2"}},
		&naming.Instance{Addrs: []string{"http://127.0.0.5:8080", "grpc://127.0.0.5:9090"}, Metadata: map[string]string{naming.MetaCluster: "c3"}},
		&naming.Instance{Addrs: []string{"http://127.0.0.5:8080", "grpc://127.0.0.5:9090"}, Metadata: map[string]string{naming.MetaCluster: "c4"}},
	}
	ins["sh002"] = []*naming.Instance{
		&naming.Instance{Addrs: []string{"http://127.0.0.1:8080", "grpc://127.0.0.1:9090"}},
		&naming.Instance{Addrs: []string{"http://127.0.0.2:8080", "grpc://127.0.0.2:9090"}},
		&naming.Instance{Addrs: []string{"http://127.0.0.3:8080", "grpc://127.0.0.3:9090"}},
	}
	insInfo = &naming.InstancesInfo{
		Instances: ins,
	}
	ok = true
	return
}

func (m *mockResolver) Watch() <-chan struct{} {
	event := make(chan struct{}, 10)
	event <- struct{}{}
	event <- struct{}{}
	event <- struct{}{}
	return event
}

func (m *mockResolver) Close() (err error) {
	return
}

type mockBuilder struct {
}

func (m *mockBuilder) Build(id string) naming.Resolver {
	return &mockResolver{}
}

func (m *mockBuilder) Scheme() string {
	return "discovery"
}

type mockClientConn struct {
	mu    sync.Mutex
	addrs []resolver.Address
}

func (m *mockClientConn) NewAddress(addresses []resolver.Address) {
	m.mu.Lock()
	m.addrs = addresses
	m.mu.Unlock()
}
func (m *mockClientConn) NewServiceConfig(serviceConfig string) {

}
func (m *mockClientConn) UpdateState(state resolver.State) {

}
func TestBuilder(t *testing.T) {
	target := resolver.Target{Endpoint: "discovery://default/service.name?zone=sh001&cluster=c1&cluster=c2&cluster=c3"}
	mb := &mockBuilder{}
	b := &Builder{mb}
	cc := &mockClientConn{}
	r, err := b.Build(target, cc, resolver.BuildOption{})
	if err != nil {
		t.Fatal(err)
	}
	res := r.(*Resolver)
	if res.zone != "sh001" {
		t.Fatalf("want sh001, but got:%s", res.zone)
	}
	if len(res.clusters) != 3 {
		t.Fatalf("want c1,c2,c3, but got:%v", res.clusters)
	}
	time.Sleep(time.Second)
	cc.mu.Lock()
	defer cc.mu.Unlock()
	if len(cc.addrs) != 5 {
		t.Fatalf("want 3, but got:%v", cc.addrs)
	}
}
