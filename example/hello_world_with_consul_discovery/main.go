package main

import (
	"fmt"
	"log"

	"github.com/minus5/nsqm"
	"github.com/minus5/nsqm/discovery/consul"
	nsq "github.com/nsqio/go-nsq"
)

const (
	topic   = "hello_world"
	channel = "app"
)

var msgs = make(chan string)

func main() {
	// discovery
	dcy, err := consul.New("127.0.0.1:8500")
	if err != nil {
		log.Fatal(err)
	}
	// show discovered configuration
	fmt.Print("NSQLookupdAddresses: ")
	fmt.Println(dcy.NSQLookupdAddresses())
	fmt.Printf("NSQDAddress: ")
	fmt.Println(dcy.NSQDAddress())

	//dcy.Monitor()

	// configuration with discovery
	cfgr := nsqm.WithDiscovery(dcy)
	// create producer
	producer, err := nsqm.NewProducer(cfgr)
	if err != nil {
		log.Fatal(err)
	}
	// create consumer
	consumer, err := nsqm.NewConsumer(cfgr, topic, channel, &handler{})
	if err != nil {
		log.Fatal(err)
	}
	// publish a message
	if err := producer.Publish(topic, []byte("Hello World")); err != nil {
		log.Fatal(err)
	}
	// waith for consumer to receive a message
	fmt.Printf("received: %s\n", <-msgs)

	// for {
	// 	fmt.Printf("received: %s\n", <-msgs)
	// }
	producer.Stop()
	consumer.Stop()
}

type handler struct{}

func (h *handler) HandleMessage(m *nsq.Message) error {
	msgs <- string(m.Body)
	return nil
}
