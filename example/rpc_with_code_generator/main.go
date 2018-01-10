package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/minus5/nsqm"
	"github.com/minus5/nsqm/discovery/consul"
	"github.com/minus5/nsqm/example/rpc_with_code_generator/service"
	"github.com/minus5/nsqm/example/rpc_with_code_generator/service/api"
	"github.com/minus5/nsqm/example/rpc_with_code_generator/service/api/nsq"
)

func consulConfig() *nsqm.Config {
	logFlags := log.LstdFlags | log.Lmicroseconds | log.Lshortfile
	log.SetFlags(logFlags)
	dcy, err := consul.Local()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := nsqm.WithDiscovery(dcy)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Logger = log.New(os.Stderr, "", logFlags)
	cfg.LogLevel = 2
	return cfg
}

func main() {
	// read useConsul command line switch
	var useConsul bool
	flag.BoolVar(&useConsul, "consul", false, "use consul for service discovery")
	flag.Parse()

	// get configuration
	var cfg *nsqm.Config
	if useConsul {
		cfg = consulConfig() // use consul stack
	} else {
		cfg = nsqm.Local() // use only local nsqd
	}
	// tests are faster with this setting !?
	cfg.NSQConfig.MaxInFlight = 1

	// start server
	srv, err := nsq.Server(cfg, service.New())
	if err != nil {
		log.Fatal(err)
	}
	defer srv.Close()

	// run client
	client(cfg)
}

func client(cfg *nsqm.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	c, err := nsq.Client(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	defer cancel()

	add := func(x, y int) {
		fmt.Printf("%d + %d = ", x, y)
		rsp, err := c.Add(ctx, api.TwoReq{X: x, Y: y})
		showError(err)
		if err == nil {
			fmt.Printf("%d\n", rsp.Z)
		}
	}
	multiply := func(x, y int) {
		fmt.Printf("%d * %d = ", x, y)
		rsp, err := c.Multiply(ctx, api.TwoReq{X: x, Y: y})
		showError(err)
		if err == nil {
			fmt.Printf("%d\n", rsp.Z)
		}
	}
	cube := func(x int) {
		fmt.Printf("%d^2 = ", x)
		rsp, err := c.Cube(ctx, x)
		showError(err)
		if err == nil {
			fmt.Printf("%d\n", *rsp)
		}
	}

	add(2, 3)
	add(128, 129)
	multiply(2, 3)
	multiply(64, 3)
	cube(12)
	cube(15)
}

// showError example of transfering typed application errors from server to client
func showError(err error) {
	if err == nil {
		return
	}
	if err == api.Overflow {
		fmt.Printf("%s\n", err)
		return
	}
	if err == context.Canceled || err == context.DeadlineExceeded {
		fmt.Printf("%s\n", err)
		return
	}
	panic(err)
}
