package model

// ArgRegister define register param.
type ArgRegister struct {
	Region          string
	Zone            string
	Env             string
	Appid           string
	Treeid          int64
	Hostname        string
	Status          uint32
	HTTP            string
	RPC             string
	Color           string
	Weight          int
	Version         string
	Metadata        string
	Replication     bool
	LatestTimestamp int64
	DirtyTimestamp  int64
}

// ArgRenew define renew params.
type ArgRenew struct {
	Region         string `form:"region" validate:"required"`
	Zone           string `form:"zone" validate:"required"`
	Env            string `form:"env" validate:"required"`
	Appid          string `form:"appid" validate:"required"`
	Treeid         int64  `form:"treeid"`
	Hostname       string `form:"hostname" validate:"required"`
	Replication    bool   `form:"replication"`
	DirtyTimestamp int64  `form:"dirty_timestamp"`
}

// ArgCancel define cancel params.
type ArgCancel struct {
	Region          string `form:"region" validate:"required"`
	Zone            string `form:"zone" validate:"required"`
	Env             string `form:"env" validate:"required"`
	Appid           string `form:"appid" validate:"required"`
	Treeid          int64  `form:"treeid"`
	Hostname        string `form:"hostname" validate:"required"`
	Replication     bool   `form:"replication"`
	LatestTimestamp int64  `form:"latest_timestamp"`
}

// ArgFetch define fetch param.
type ArgFetch struct {
	Region string `form:"region" validate:"required"`
	Zone   string `form:"zone" validate:"required"`
	Env    string `form:"env" validate:"required"`
	Appid  string `form:"appid"`
	Treeid int64  `form:"treeid"`
	Status uint32 `form:"status" validate:"required"`
}

// ArgPoll define poll param.
type ArgPoll struct {
	Region          string `form:"region" validate:"required"`
	Zone            string `form:"zone" validate:"required"`
	Env             string `form:"env" validate:"required"`
	Appid           string `form:"appid"`
	Treeid          int64  `form:"treeid"`
	Hostname        string `form:"hostname" validate:"required"`
	LatestTimestamp int64  `form:"latest_timestamp"`
}

// ArgPolls define poll param.
type ArgPolls struct {
	Region          string   `form:"region" validate:"required"`
	Zone            string   `form:"zone" validate:"required"`
	Env             string   `form:"env" validate:"required"`
	Appid           []string `form:"appid,split"`
	Treeid          []int64  `form:"treeid,split" validate:"required"`
	Hostname        string   `form:"hostname,split" validate:"required"`
	LatestTimestamp []int64  `form:"latest_timestamp,split"`
}

// ArgSet define set param.
type ArgSet struct {
	Region       string   `form:"region" validate:"required"`
	Zone         string   `form:"zone" validate:"required"`
	Env          string   `form:"env" validate:"required"`
	Appid        string   `form:"appid" validate:"required"`
	Hostname     []string `form:"hostname,split"`
	Weight       []int64  `form:"weight,split"`
	Status       []int64  `form:"status,split"`
	Color        []string `form:"color,split"`
	Replication  bool     `form:"replication"`
	SetTimestamp int64    `form:"set_timestamp"`
}
