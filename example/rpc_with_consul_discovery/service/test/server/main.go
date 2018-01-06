package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/minus5/nsqm/example/rpc_with_consul_discovery/service"
	"github.com/minus5/nsqm/example/rpc_with_consul_discovery/service/api/nsq"
)

func main() {
	srv, err := nsq.Server(service.New())
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
