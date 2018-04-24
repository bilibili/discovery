package model

// ArgRegister define register param.
type ArgRegister struct {
	Zone            string   `form:"zone" validate:"required"`
	Env             string   `form:"env" validate:"required"`
	AppID           string   `form:"appid" validate:"required"`
	Hostname        string   `form:"hostname" validate:"required"`
	Status          uint32   `form:"status" validate:"required"`
	Addrs           []string `form:"addrs"`
	Color           string   `form:"color"`
	Version         string   `form:"version"`
	Metadata        string   `form:"metadata"`
	Replication     bool     `form:"replication"`
	LatestTimestamp int64    `form:"latest_timestamp"`
	DirtyTimestamp  int64    `form:"dirty_timestamp"`
}

// ArgRenew define renew params.
type ArgRenew struct {
	Zone           string `form:"zone" validate:"required"`
	Env            string `form:"env" validate:"required"`
	AppID          string `form:"appid" validate:"required"`
	Hostname       string `form:"hostname" validate:"required"`
	Replication    bool   `form:"replication"`
	DirtyTimestamp int64  `form:"dirty_timestamp"`
}

// ArgCancel define cancel params.
type ArgCancel struct {
	Zone            string `form:"zone" validate:"required"`
	Env             string `form:"env" validate:"required"`
	AppID           string `form:"appid" validate:"required"`
	Hostname        string `form:"hostname" validate:"required"`
	Replication     bool   `form:"replication"`
	LatestTimestamp int64  `form:"latest_timestamp"`
}

// ArgFetch define fetch param.
type ArgFetch struct {
	Zone   string `form:"zone" validate:"required"`
	Env    string `form:"env" validate:"required"`
	AppID  string `form:"appid"`
	Status uint32 `form:"status" validate:"required"`
}

// ArgFetchs define fetchs arg.
type ArgFetchs struct {
	Zone   string   `form:"zone"`
	Env    string   `form:"env" validate:"required"`
	AppID  []string `form:"appid,split"`
	Status uint32   `form:"status" validate:"required"`
}

// ArgPoll define poll param.
type ArgPoll struct {
	Zone            string `form:"zone" validate:"required"`
	Env             string `form:"env" validate:"required"`
	AppID           string `form:"appid"`
	Hostname        string `form:"hostname" validate:"required"`
	LatestTimestamp int64  `form:"latest_timestamp"`
}

// ArgPolls define poll param.
type ArgPolls struct {
	Zone            string   `form:"zone" validate:"required"`
	Env             string   `form:"env" validate:"required"`
	AppID           []string `form:"appid"`
	Hostname        string   `form:"hostname" validate:"required"`
	LatestTimestamp []int64  `form:"latest_timestamp"`
}

// ArgSet define set param.
type ArgSet struct {
	Zone         string   `form:"zone" validate:"required"`
	Env          string   `form:"env" validate:"required"`
	AppID        string   `form:"appid" validate:"required"`
	Hostname     []string `form:"hostname"`
	Status       []int64  `form:"status"`
	Color        []string `form:"color"`
	Replication  bool     `form:"replication"`
	SetTimestamp int64    `form:"set_timestamp"`
}
