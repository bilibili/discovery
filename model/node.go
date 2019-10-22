package model

import (
	"encoding/json"
)

// NodeStatus Status of instance
type NodeStatus int

const (
	// NodeStatusUP Ready to receive register
	NodeStatusUP NodeStatus = iota
	// NodeStatusLost lost with each other
	NodeStatusLost
)

const (
	// AppID is discvoery id
	AppID = "infra.discovery"
)

// Node node
type Node struct {
	Addr   string     `json:"addr"`
	Status NodeStatus `json:"status"`
	Zone   string     `json:"zone"`
}

// Scheduler info.
type Scheduler struct {
	AppID   string                   `json:"app_id,omitempty"`
	Env     string                   `json:"env,omitempty"`
	Clients map[string]*ZoneStrategy `json:"clients"` // zone-ratio
	Remark  string                   `json:"remark"`
}

// ZoneStrategy is the scheduling strategy of all zones
type ZoneStrategy struct {
	Zones map[string]*Strategy `json:"zones"`
}

// Strategy is zone scheduling strategy.
type Strategy struct {
	Weight int64 `json:"weight"`
}

// Zone info.
type Zone struct {
	Src string         `json:"src"`
	Dst map[string]int `json:"dst"`
}

// Set Scheduler conf settter.
func (s *Scheduler) Set(content string) (err error) {
	if err = json.Unmarshal([]byte(content), &s); err != nil {
		return
	}
	return
}
