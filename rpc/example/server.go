// +build server

package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/minus5/nsqm/rpc"
	"github.com/minus5/svckit/signal"
	"github.com/nsqio/go-nsq"
)

func main() {
	nsqdTCPAddr := "127.0.0.1:4150"
	cfg := nsq.NewConfig()
	producer, err := nsq.NewProducer(nsqdTCPAddr, cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	srv := &server{}
	transport := rpc.NewServer(ctx, srv, producer)

	cfg = nsq.NewConfig()
	reqTopic := "request"
	channel := "server"
	consumer, err := nsq.NewConsumer(reqTopic, channel, cfg)
	if err != nil {
		log.Fatal(err)
	}
	consumer.AddConcurrentHandlers(transport, 256)
	err = consumer.ConnectToNSQD(nsqdTCPAddr)
	if err != nil {
		log.Fatal(err)
	}

	defer producer.Stop()
	defer cancel()
	defer consumer.Stop()

	//p.SetLogger(defaults.logger, defaults.logLevel)
	//return &Producer{nsqProducer: p, topic: topic}, nil
	signal.WaitForInterupt()
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
