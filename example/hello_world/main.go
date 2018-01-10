package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/minus5/nsqm"
	"github.com/minus5/nsqm/discovery/consul"
	nsq "github.com/nsqio/go-nsq"
)

const (
	topic   = "hello_world"
	channel = "app"
)

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

	// create producer
	producer, err := nsqm.NewProducer(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// create consumer
	hnd := &handler{msgs: make(chan string)}
	consumer, err := nsqm.NewConsumer(cfg, topic, channel, hnd)
	if err != nil {
		log.Fatal(err)
	}

	// send a message with producer
	msg := time.Now().Format(time.RFC3339Nano)
	fmt.Printf("sending : %s\n", msg)
	producer.Publish(topic, []byte(msg))

	// wait for consumer to receive a message
	fmt.Printf("received: %s\n", <-hnd.msgs)

	// cleanup
	producer.Stop()
	consumer.Stop()
}

type handler struct {
	msgs chan string
}

func (h *handler) HandleMessage(m *nsq.Message) error {
	h.msgs <- string(m.Body)
	return nil
}

func consulConfig() *nsqm.Config {
	// get consul discovery
	dcy, err := consul.Local()
	if err != nil {
		log.Fatal(err)
	}
	// show discovered configuration
	la, _ := dcy.NSQLookupdAddresses()
	na, _ := dcy.NSQDAddress()
	fmt.Printf("config from consul:\n\tnsqd: %s,\n\tnsqlookupds:%v\n", na, la)
	// create configuration from discovery
	cfg, err := nsqm.WithDiscovery(dcy)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}
