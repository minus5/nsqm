// +build client

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/minus5/nsqm"
	"github.com/minus5/nsqm/rpc"
)

const (
	reqTopic = "request"  // topic for sending request to server
	rspTopic = "response" // topic for getting responses from server
	channel  = "client"   // channel name for rspTopic topic
)

func main() {
	// configuration
	cfgr := nsqm.Local()

	// nsq producer for sending requests
	producer, err := nsqm.NewProducer(cfgr)
	if err != nil {
		log.Fatal(err)
	}

	// rpc client: sends requests, waits and accepts responses
	//             provides interface for application
	rpcClient := rpc.NewClient(producer, reqTopic, rspTopic)

	// create consumer arround rpcClient
	consumer, err := nsqm.NewConsumer(cfgr, rspTopic, channel, rpcClient)
	if err != nil {
		log.Fatal(err)
	}

	// application client
	client := &client{t: rpcClient}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// clean exit
	defer producer.Stop() // 3. stop producing new requests
	defer cancel()        // 2. cancel any pending (waiting for responses)
	defer consumer.Stop() // 1. stop listening for responses

	x := 2
	y := 3
	z, err := client.Add(ctx, x, y)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d + %d =  %d\n", x, y, z)
}

// transport is application interface for sending request to the remote server
// method - server method name
// req    - request
// returns:
//   reponse
//   application error, string, "" if there is no error
//   transport error
type transport interface {
	Call(ctx context.Context, method string, req []byte) ([]byte, string, error)
}

type client struct {
	t transport
}

func (c *client) Add(ctx context.Context, x, y int) (int, error) {
	req := &request{X: x, Y: y}
	// marshall request
	reqBuf, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}
	// pass request to trasport, and get response
	rspBuf, _, err := c.t.Call(ctx, "Add", reqBuf)
	// request was canceled
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	// request failed
	if err != nil {
		return 0, err
	}
	// unmarshal response
	var rsp response
	err = json.Unmarshal(rspBuf, &rsp)
	if err != nil {
		return 0, err
	}
	return rsp.Z, nil
}

// dto structures

type request struct {
	X int
	Y int
}

type response struct {
	Z int
}
