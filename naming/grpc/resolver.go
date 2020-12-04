package resolver

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/bilibili/discovery/naming"

	log "github.com/go-kratos/kratos/pkg/log"
	"google.golang.org/grpc/resolver"
)

const (
	// Scheme is the scheme of discovery address
	Scheme = "grpc"
)

var (
	_ resolver.Resolver = &Resolver{}
	_ resolver.Builder  = &Builder{}
)

// MD is context metadata for balancer and resolver.
type MD struct {
	Weight int64
	Color  string
}

// Register register resolver builder if nil.
func Register(b naming.Builder) {
	resolver.Register(&Builder{b})
}

// Builder is also a resolver builder.
// It's build() function always returns itself.
type Builder struct {
	naming.Builder
}

// Build returns itself for Resolver, because it's both a builder and a resolver.
func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	// discovery://default/service.name?zone=sh001&cluster=c1&cluster=c2&cluster=c3
	dsn := strings.SplitN(target.Endpoint, "?", 2)
	if len(dsn) == 0 {
		return nil, fmt.Errorf("grpc resolver: parse target.Endpoint(%s) failed! the endpoint is empty", target.Endpoint)
	}
	// parse params info
	zone := os.Getenv("ZONE")
	clusters := map[string]struct{}{}
	if len(dsn) == 2 {
		if u, err := url.ParseQuery(dsn[1]); err == nil {
			if zones := u[naming.MetaZone]; len(zones) > 0 {
				zone = zones[0]
			}
			for _, c := range u[naming.MetaCluster] {
				clusters[c] = struct{}{}
			}
		}
	}
	r := &Resolver{
		cc:       cc,
		nr:       b.Builder.Build(dsn[0]),
		quit:     make(chan struct{}, 1),
		zone:     zone,
		clusters: clusters,
	}
	go r.watcher()
	return r, nil
}

// Resolver watches for the updates on the specified target.
// Updates include address updates and service config updates.
type Resolver struct {
	nr   naming.Resolver
	cc   resolver.ClientConn
	quit chan struct{}

	zone     string
	clusters map[string]struct{}
}

// Close is a noop for Resolver.
func (r *Resolver) Close() {
	select {
	case r.quit <- struct{}{}:
		r.nr.Close()
	default:
	}
}

// ResolveNow is a noop for Resolver.
func (r *Resolver) ResolveNow(o resolver.ResolveNowOptions) {
}

func (r *Resolver) watcher() {
	event := r.nr.Watch()
	for {
		select {
		case <-r.quit:
			return
		case _, ok := <-event:
			if !ok {
				return
			}
		}
		ins, ok := r.nr.Fetch()
		if ok {
			instances, ok := ins.Instances[r.zone]
			if !ok {
				for _, value := range ins.Instances {
					instances = append(instances, value...)
				}
			}
			if len(instances) > 0 {
				r.newAddress(instances)
			}
		}
	}
}

func (r *Resolver) newAddress(instances []*naming.Instance) {
	var (
		totalWeight int64
		addrs       = make([]resolver.Address, 0, len(instances))
	)
	for n, ins := range instances {
		if len(r.clusters) > 0 {
			if _, ok := r.clusters[ins.Metadata[naming.MetaCluster]]; !ok {
				continue
			}
		}
		rpcAddr, color, weight := extractAddrs(ins)
		if rpcAddr == "" {
			log.Warn("grpc resolver: invalid rpc address(%s,%s,%v) found!", ins.AppID, ins.Hostname, ins.Addrs)
			continue
		}
		if weight <= 0 {
			if totalWeight == 0 {
				weight = 10
			} else {
				weight = totalWeight / int64(n)
			}
		}
		totalWeight += weight
		addr := resolver.Address{
			Addr:       rpcAddr,
			Type:       resolver.Backend,
			ServerName: ins.AppID,
			Metadata:   &MD{Weight: weight, Color: color},
		}
		addrs = append(addrs, addr)
	}
	r.cc.NewAddress(addrs)
}

func extractAddrs(ins *naming.Instance) (addr, color string, weight int64) {
	color = ins.Metadata[naming.MetaColor]
	weight, _ = strconv.ParseInt(ins.Metadata[naming.MetaWeight], 10, 64)
	for _, a := range ins.Addrs {
		u, err := url.Parse(a)
		if err == nil && u.Scheme == Scheme {
			addr = u.Host
		}
	}
	return
}
