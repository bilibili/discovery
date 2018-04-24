package conf

import (
	"flag"
	"time"

	"github.com/Bilibili/discovery/lib/http"
	"github.com/BurntSushi/toml"
)

var (
	confPath string
	// Conf conf
	Conf = &Config{}
	// ConfCh for update node of server.
	ConfCh = make(chan struct{}, 1)
)

// Config config
type Config struct {
	Nodes      []string
	Zones      map[string]string // addr => name
	HTTPServer *ServerConfig
	HTTPClient *http.ClientConfig
}

// ServerConfig Http Servers conf.
type ServerConfig struct {
	Addr         string
	MaxListen    int32
	Timeout      time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func init() {
	// flag.StringVar(&confPath, "conf", "discovery-example.toml", "config path")
	flag.StringVar(&confPath, "conf", "", "config path")
}

// Init init conf
func Init() (err error) {
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}
