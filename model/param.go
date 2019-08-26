package model

// ArgRegister define register param.
type ArgRegister struct {
	Region          string   `form:"region"`
	Zone            string   `form:"zone" validate:"required"`
	Env             string   `form:"env" validate:"required"`
	AppID           string   `form:"appid" validate:"required"`
	Hostname        string   `form:"hostname" validate:"required"`
	Status          uint32   `form:"status" validate:"required"`
	Addrs           []string `form:"addrs" validate:"gt=0"`
	Version         string   `form:"version"`
	Metadata        string   `form:"metadata"`
	Replication     bool     `form:"replication"`
	LatestTimestamp int64    `form:"latest_timestamp"`
	DirtyTimestamp  int64    `form:"dirty_timestamp"`
	FromZone        bool     `form:"from_zone"`
}

// ArgRenew define renew params.
type ArgRenew struct {
	Zone           string `form:"zone" validate:"required"`
	Env            string `form:"env" validate:"required"`
	AppID          string `form:"appid" validate:"required"`
	Hostname       string `form:"hostname" validate:"required"`
	Replication    bool   `form:"replication"`
	DirtyTimestamp int64  `form:"dirty_timestamp"`
	FromZone       bool   `form:"from_zone"`
}

// ArgCancel define cancel params.
type ArgCancel struct {
	Zone            string `form:"zone" validate:"required"`
	Env             string `form:"env" validate:"required"`
	AppID           string `form:"appid" validate:"required"`
	Hostname        string `form:"hostname" validate:"required"`
	FromZone        bool   `form:"from_zone"`
	Replication     bool   `form:"replication"`
	LatestTimestamp int64  `form:"latest_timestamp"`
}

// ArgFetch define fetch param.
type ArgFetch struct {
	Zone   string `form:"zone"`
	Env    string `form:"env" validate:"required"`
	AppID  string `form:"appid" validate:"required"`
	Status uint32 `form:"status" validate:"required"`
}

// ArgFetchs define fetchs arg.
type ArgFetchs struct {
	Zone   string   `form:"zone"`
	Env    string   `form:"env" validate:"required"`
	AppID  []string `form:"appid" validate:"gt=0"`
	Status uint32   `form:"status" validate:"required"`
}

// ArgPoll define poll param.
type ArgPoll struct {
	Zone            string `form:"zone"`
	Env             string `form:"env" validate:"required"`
	AppID           string `form:"appid" validate:"required"`
	Hostname        string `form:"hostname" validate:"required"`
	LatestTimestamp int64  `form:"latest_timestamp"`
}

// ArgPolls define poll param.
type ArgPolls struct {
	Zone            string   `form:"zone"`
	Env             string   `form:"env" validate:"required"`
	AppID           []string `form:"appid" validate:"gt=0"`
	Hostname        string   `form:"hostname" validate:"required"`
	LatestTimestamp []int64  `form:"latest_timestamp"`
}

// ArgSet define set param.
type ArgSet struct {
	Region       string   `form:"region"`
	Zone         string   `form:"zone" validate:"required"`
	Env          string   `form:"env" validate:"required"`
	AppID        string   `form:"appid" validate:"required"`
	Hostname     []string `form:"hostname" validate:"gte=0"`
	Status       []int64  `form:"status" validate:"gte=0"`
	Metadata     []string `form:"metadata" validate:"gte=0"`
	Replication  bool     `form:"replication"`
	FromZone     bool     `form:"from_zone"`
	SetTimestamp int64    `form:"set_timestamp"`
}
