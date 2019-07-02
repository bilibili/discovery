package conf

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	log "github.com/bilibili/kratos/pkg/log"
	http "github.com/bilibili/kratos/pkg/net/http/blademaster"
)

var (
	confPath      string
	schedulerPath string
	region        string
	zone          string
	deployEnv     string
	hostname      string
	// Conf conf
	Conf = &Config{}
)

func init() {
	var err error
	if hostname, err = os.Hostname(); err != nil || hostname == "" {
		hostname = os.Getenv("HOSTNAME")
	}
	flag.StringVar(&confPath, "conf", "discovery-example.toml", "config path")
	flag.StringVar(&hostname, "hostname", hostname, "machine hostname")
	flag.StringVar(&schedulerPath, "scheduler", "scheduler.json", "scheduler info")
}

// Config config.
type Config struct {
	Nodes      []string
	Zones      map[string][]string
	HTTPServer *http.ServerConfig
	HTTPClient *http.ClientConfig
	Env        *Env
	Log        *log.Config
	Scheduler  []byte
}

// Fix fix env config.
func (c *Config) Fix() (err error) {
	if c.Env == nil {
		c.Env = new(Env)
	}
	if c.Env.Region == "" {
		c.Env.Region = region
	}
	if c.Env.Zone == "" {
		c.Env.Zone = zone
	}
	if c.Env.Host == "" {
		c.Env.Host = hostname
	}
	if c.Env.DeployEnv == "" {
		c.Env.DeployEnv = deployEnv
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

// Init init conf
func Init() (err error) {
	if _, err = toml.DecodeFile(confPath, &Conf); err != nil {
		return
	}
	if schedulerPath != "" {
		Conf.Scheduler, _ = ioutil.ReadFile(schedulerPath)
	}
	return Conf.Fix()
}
