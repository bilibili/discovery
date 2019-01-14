package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"discovery/conf"
	"discovery/discovery"
	"discovery/http"

	log "github.com/golang/glog"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Errorf("conf.Init() error(%v)", err)
		panic(err)
	}
	dis, cancel := discovery.New(conf.Conf)
	http.Init(conf.Conf, dis)
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("discovery get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			cancel()
			time.Sleep(time.Second)
			log.Info("discovery quit !!!")
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
