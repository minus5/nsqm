package main

import (
	"context"
	"fmt"
	"log"

	"github.com/minus5/nsqm/example/rpc_with_consul_discovery/service"
	"github.com/minus5/nsqm/example/rpc_with_consul_discovery/service/api"
	"github.com/minus5/nsqm/example/rpc_with_consul_discovery/service/api/nsq"
)

func main() {
	srv, err := nsq.Server(service.New())
	if err != nil {
		log.Fatal(err)
	}
	defer srv.Close()

	client()
}

func client() {
	ctx, cancel := context.WithCancel(context.Background())
	c, err := nsq.Client()
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
