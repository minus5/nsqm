// +build client

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/minus5/nsqm"
	"github.com/minus5/nsqm/discovery/consul"
	"github.com/minus5/nsqm/example/rpc_with_code_generator/service/api"
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

	ctx, cancel := context.WithCancel(context.Background())
	c, err := nsq.Client(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	defer cancel()

	add := func(x, y int) {
		fmt.Printf("%d + %d = ", x, y)
		rsp, err := c.Add(ctx, api.TwoReq{X: x, Y: y}, nil)
		showError(err)
		if err == nil {
			fmt.Printf("%d\n", rsp.Z)
		}
	}
	multiply := func(x, y int) {
		fmt.Printf("%d * %d = ", x, y)
		rsp, err := c.Multiply(ctx, api.TwoReq{X: x, Y: y}, nil)
		showError(err)
		if err == nil {
			fmt.Printf("%d\n", rsp.Z)
		}
	}
	cube := func(x int) {
		fmt.Printf("%d^2 = ", x)
		rsp, err := c.Cube(ctx, x, nil)
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
