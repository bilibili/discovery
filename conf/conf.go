package conf

import (
	"flag"

	"github.com/Bilibili/discovery/lib/http"
	"github.com/Bilibili/discovery/lib/time"
	"github.com/BurntSushi/toml"
)

var (
	confPath string
	// Conf conf
	Conf = &Config{}
)

// Config config
type Config struct {
	Zone       string
	Nodes      []string
	Zones      map[string]string // addr => name
	HTTPServer *ServerConfig
	HTTPClient *http.ClientConfig
}

// ServerConfig Http Servers conf.
type ServerConfig struct {
	Addr         string
	Timeout      time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func init() {
	flag.StringVar(&confPath, "conf", "discovery-example.toml", "config path")
}

// Init init conf
func Init() (err error) {
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}
