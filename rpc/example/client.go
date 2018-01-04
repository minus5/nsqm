package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/minus5/nsqm/rpc"
	nsq "github.com/nsqio/go-nsq"
)

func main() {
	nsqdTCPAddr := "127.0.0.1:4150"
	cfg := nsq.NewConfig()
	producer, err := nsq.NewProducer(nsqdTCPAddr, cfg)
	if err != nil {
		log.Fatal(err)
	}

	reqTopic := "request"
	rspTopic := "response"
	channel := "client"
	cfg = nsq.NewConfig()
	consumer, err := nsq.NewConsumer(rspTopic, channel, cfg)
	if err != nil {
		log.Fatal(err)
	}
	transport := rpc.NewClient(producer, consumer, reqTopic, rspTopic)
	consumer.AddConcurrentHandlers(transport, 256)
	err = consumer.ConnectToNSQD(nsqdTCPAddr)
	if err != nil {
		log.Fatal(err)
	}

	client := &client{t: transport}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer producer.Stop()
	defer cancel()
	defer consumer.Stop()

	x := 2
	y := 3
	z, err := client.Add(ctx, x, y)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d + %d =  %d\n", x, y, z)
}

type transport interface {
	Call(ctx context.Context, typ string, req []byte) ([]byte, string, error)
}

type client struct {
	t transport
}

func (c *client) Add(ctx context.Context, x, y int) (int, error) {
	req := &request{X: x, Y: y}
	reqBuf, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}
	rspBuf, _, err := c.t.Call(ctx, "Add", reqBuf)
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	if err != nil {
		return 0, err
	}
	var rsp response
	err = json.Unmarshal(rspBuf, &rsp)
	if err != nil {
		return 0, err
	}
	return rsp.Z, nil
}

type request struct {
	X int
	Y int
}

type response struct {
	Z int
}
