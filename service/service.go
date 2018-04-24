package service

import (
	"os"
	"sync"

	"github.com/Bilibili/discovery/conf"
	"github.com/Bilibili/discovery/dao"
	"github.com/Bilibili/discovery/lib/http"
)

// Service discovery main service
type Service struct {
	c        *conf.Config
	client   *http.Client
	registry *dao.Registry
	nodes    *dao.Nodes
	tLock    sync.RWMutex
	tree     map[int64]string // treeid->appid
	env      *env
}

type env struct {
	Region string
	Zone   string
}

// New get a discovery service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:        c,
		client:   http.NewClient(c.HTTPClient),
		registry: dao.NewRegistry(),
		nodes:    dao.NewNodes(c),
		tree:     make(map[int64]string),
	}
	s.getEnv()
	s.syncUp()
	go s.updateNodes()
	return
}

func (s *Service) getEnv() {
	s.env = &env{
		Region: os.Getenv("REGION"),
		Zone:   os.Getenv("ZONE"),
	}
}
