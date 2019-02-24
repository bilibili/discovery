package model

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
	AppID  string `json:"app_id,omitempty"`
	Env    string `json:"env"`
	Zones  []Zone `json:"zones"` // zone-ratio
	Remark string `json:"remark"`
}

// Zone info.
type Zone struct {
	Src string         `json:"src"`
	Dst map[string]int `json:"dst"`
}
