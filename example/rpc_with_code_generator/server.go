// +build server

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/minus5/nsqm"
	"github.com/minus5/nsqm/discovery/consul"
	"github.com/minus5/nsqm/example/rpc_with_code_generator/service"
	"github.com/minus5/nsqm/example/rpc_with_code_generator/service/api/nsq"
)

func consulConfig() *nsqm.Config {
	dcy, err := consul.Local()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := nsqm.WithDiscovery(dcy)
	if err != nil {
		log.Fatal(err)
	}
	cfg.NSQConfig.MaxInFlight = 1
	return cfg
}

func main() {
	cfg := consulConfig()

	srv, err := nsq.Server(cfg, service.New())
	if err != nil {
		log.Fatal(err)
	}
	defer srv.Close()

	waitForInterupt()
}

func waitForInterupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}
