package conf

import (
	"flag"
	"os"

	"github.com/Bilibili/discovery/lib/http"
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
	Env        *Env
}

// Fix fix env config.
func (c *Config) Fix() (err error) {
	if c.Env == nil {
		c.Env = new(Env)
	}
	if c.Env.Region == "" {
		c.Env.Region = os.Getenv("REGION")
	}
	if c.Env.Zone == "" {
		c.Env.Zone = os.Getenv("ZONE")
	}
	if c.Env.Host == "" {
		c.Env.Host, _ = os.Hostname()
	}
	if c.Env.DeployEnv == "" {
		c.Env.DeployEnv = os.Getenv("DEPLOY_ENV")
	}
	// FIXME use env instead of zone
	if c.Zone != "" {
		c.Env.Zone = c.Zone
	} else {
		c.Zone = c.Env.Zone
	}
	return
}

// Env is disocvery env.
type Env struct {
	Region    string
	Zone      string
	Host      string
	DeployEnv string
}

// ServerConfig Http Servers conf.
type ServerConfig struct {
	Addr string
}

func init() {
	flag.StringVar(&confPath, "conf", "discovery-example.toml", "config path")
}

// Init init conf
func Init() (err error) {
	if _, err = toml.DecodeFile(confPath, &Conf); err != nil {
		return
	}
	return Conf.Fix()
}
