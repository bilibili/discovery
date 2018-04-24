package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bilibili/discovery/conf"
	"github.com/Bilibili/discovery/http"
	log "github.com/golang/glog"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Errorf("conf.Init() error(%v)", err)
		panic(err)
	}
	// http init
	http.Init(conf.Conf)
	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("discovery get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Info("discovery quit !!!")
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
