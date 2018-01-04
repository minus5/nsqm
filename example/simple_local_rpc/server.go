// +build server

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/minus5/nsqm"
	"github.com/minus5/nsqm/rpc"
)

const (
	reqTopic = "request"
	channel  = "server"
)

func main() {
	cfgr := nsqm.Local()

	producer, err := nsqm.NewProducer(cfgr)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	srv := &server{}
	transport := rpc.NewServer(ctx, srv, producer)

	consumer, err := nsqm.NewConsumer(cfgr, reqTopic, channel, transport)
	if err != nil {
		log.Fatal(err)
	}

	defer producer.Stop()
	defer cancel()
	defer consumer.Stop()

	waitForInterupt()
}

func waitForInterupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

type server struct{}

func (s *server) Serve(ctx context.Context, method string, reqBuf []byte) ([]byte, error) {
	switch method {
	case "Add":
		var req request
		err := json.Unmarshal(reqBuf, &req)
		if err != nil {
			return nil, err
		}
		z := s.add(req.X, req.Y)
		rsp := response{Z: z}
		rspBuf, err := json.Marshal(rsp)
		if err != nil {
			return nil, err
		}
		return rspBuf, nil
	default:
		return nil, nil
	}
}

func (s *server) add(x, y int) int {
	return x + y
}

type request struct {
	X int
	Y int
}

type response struct {
	Z int
}
