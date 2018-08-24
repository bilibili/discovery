package naming

import (
	"context"
)

// Instance represents a server the client connects to.
type Instance struct {
	// Zone is IDC.
	Zone string `json:"zone"`
	// Env prod/pre、uat/fat1
	Env string `json:"env"`
	// AppID is mapping servicetree appid.
	AppID string `json:"appid"`
	// Hostname is hostname from docker.
	Hostname string `json:"hostname"`
	// Addrs is the address of app instance
	// format: scheme://host
	Addrs []string `json:"addrs"`
	// Color distinguishs grey or common.
	Color string `json:"color"`
	// Version is publishing version.
	Version string `json:"version"`
	// LastTs is instance latest updated timestamp
	LastTs int64 `json:"latest_timestamp"`
	// Metadata is the information associated with Addr, which may be used
	// to make load balancing decision.
	Metadata map[string]string `json:"metadata"`
}

// Resolver resolve naming service
type Resolver interface {
	Fetch() (map[string][]*Instance, bool)
	Watch() <-chan struct{}
	Close() error
}

// Registry Register an instance and renew automatically
type Registry interface {
	Register(ins *Instance) (cancel context.CancelFunc, err error)
	Close() error
}

// Builder resolver builder.
type Builder interface {
	Build(id string) Resolver
	Scheme() string
}
