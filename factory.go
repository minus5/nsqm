package nsqm

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/minus5/nsqm/rpc"
	nsq "github.com/nsqio/go-nsq"
)

// NewProducer creates nsq nsq.Producer from Config.
func NewProducer(cfg *Config) (*nsq.Producer, error) {
	producer, err := nsq.NewProducer(cfg.NSQDAddress, cfg.nsqConfig())
	if err != nil {
		return nil, err
	}
	producer.SetLogger(cfg.Logger, cfg.LogLevel)
	return producer, nil
}

// NewConsumer creates and configures new nsq.Consumer.
func NewConsumer(cfg *Config, topic, channel string, handler nsq.Handler) (*nsq.Consumer, error) {
	consumer, err := nsq.NewConsumer(topic, channel, cfg.nsqConfig())
	if err != nil {
		return nil, err
	}
	consumer.SetLogger(cfg.Logger, cfg.LogLevel)
	consumer.AddConcurrentHandlers(handler, cfg.Concurrency)
	if addrs := cfg.NSQLookupdAddresses; addrs != nil {
		if err := consumer.ConnectToNSQLookupds(addrs); err != nil {
			return nil, err
		}
	} else {
		if err := consumer.ConnectToNSQD(cfg.NSQDAddress); err != nil {
			return nil, err
		}
	}
	cfg.Subscribe(consumer)
	return consumer, nil
}

func NewRpcClient(cfg *Config, reqTopic string) (*RpcClient, error) {
	factoryMutex.Lock()
	defer factoryMutex.Unlock()
	// ensuring that there is only one client handler per application
	if rpcHandler != nil {
		return &RpcClient{
			reqTopic: reqTopic,
			handler:  rpcHandler}, nil
	}

	channel := appName()
	rspTopic := fmt.Sprintf("z...rsp-%s-%s", appName(), cfg.NodeName)
	producer, err := NewProducer(cfg)
	if err != nil {
		return nil, err
	}
	rpcHandler = rpc.NewClient(producer, "", rspTopic)
	consumer, err := NewConsumer(cfg, rspTopic, channel, rpcHandler)
	if err != nil {
		return nil, err
	}
	return &RpcClient{
		reqTopic: reqTopic,
		producer: producer,
		consumer: consumer,
		handler:  rpcHandler}, nil
}

type RpcClient struct {
	reqTopic string
	producer *nsq.Producer
	consumer *nsq.Consumer
	handler  *rpc.Client
}

func (c *RpcClient) Call(ctx context.Context, typ string, req []byte) ([]byte, string, error) {
	return c.handler.CallTopic(ctx, c.reqTopic, typ, req)
}

func (c *RpcClient) Close() error {
	if c.producer != nil {
		c.producer.Stop()
	}
	if c.consumer != nil {
		c.consumer.Stop()
	}
	return nil
}

var rpcHandler *rpc.Client
var factoryMutex sync.Mutex

func appName() string {
	return path.Base(os.Args[0])
}

func NewRpcServer(cfg *Config, reqTopic string, srv AppServer) (*RpcServer, error) {
	channel := appName()
	producer, err := NewProducer(cfg)
	if err != nil {
		return nil, err
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	rpcServer := rpc.NewServer(ctx, srv, producer)

	consumer, err := NewConsumer(cfg, reqTopic, channel, rpcServer)
	if err != nil {
		return nil, err
	}

	return &RpcServer{
		producer:  producer,
		ctxCancel: ctxCancel,
		consumer:  consumer,
	}, nil
}

type AppServer interface {
	Serve(ctx context.Context, typ string, req []byte) ([]byte, error)
}

type RpcServer struct {
	producer  *nsq.Producer
	ctxCancel func()
	consumer  *nsq.Consumer
}

func (s *RpcServer) Stop() {
	s.consumer.Stop() // stop receiving new requests
	s.ctxCancel()     // cancel all processing
	<-s.consumer.StopChan
}

func (s *RpcServer) Close() {
	s.producer.Stop() // stop producing responses
}
