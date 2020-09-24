package conf

import (
	"flag"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/bilibili/kratos/pkg/conf/env"
	"github.com/bilibili/kratos/pkg/conf/paladin"
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
	configKey     string
	// Conf conf
	Conf = &Config{}
)

func init() {
	var err error
	if hostname, err = os.Hostname(); err != nil || hostname == "" {
		hostname = os.Getenv("HOSTNAME")
	}
	flag.StringVar(&configKey, "confkey", "discovery-example.toml", "discovery conf key")
	flag.StringVar(&hostname, "hostname", hostname, "machine hostname")
	flag.StringVar(&schedulerPath, "scheduler", "scheduler.json", "scheduler info")
}

// Config config.
type Config struct {
	Nodes         []string
	Zones         map[string][]string
	HTTPServer    *http.ServerConfig
	HTTPClient    *http.ClientConfig
	Env           *Env
	Log           *log.Config
	Scheduler     []byte
	EnableProtect bool
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

// Env is discovery env.
type Env struct {
	Region    string
	Zone      string
	Host      string
	DeployEnv string
}

// Fix default env from kartos
func fixDefEnv() {
	region = env.Region
	zone = env.Zone
	deployEnv = env.DeployEnv
}

// Init init conf
func Init() (err error) {
	fixDefEnv()
	if err = paladin.Init(); err != nil {
		return
	}
	return paladin.Watch(configKey, Conf)
}

// Set config setter.
func (c *Config) Set(content string) (err error) {
	var tmpConf *Config
	if _, err = toml.Decode(content, &tmpConf); err != nil {
		log.Error("decode config fail %v", err)
		return
	}
	if err = tmpConf.Fix(); err != nil {
		return
	}
	*Conf = *tmpConf
	return nil
}
