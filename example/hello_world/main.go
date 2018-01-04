package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/minus5/nsqm"
	nsq "github.com/nsqio/go-nsq"
)

const (
	topic   = "hello_world"
	channel = "app"
)

var wg sync.WaitGroup

func main() {
	cfgr := nsqm.Local()

	producer, err := nsqm.NewProducer(cfgr)
	if err != nil {
		log.Fatal(err)
	}

	consumer, err := nsqm.NewConsumer(cfgr, topic, channel, &handler{})
	if err != nil {
		log.Fatal(err)
	}

	wg.Add(1)
	producer.Publish(topic, []byte("Hello World"))

	wg.Wait()
	producer.Stop()
	consumer.Stop()
}

type handler struct{}

func (h *handler) HandleMessage(m *nsq.Message) error {
	fmt.Printf("received: %s\n", m.Body)
	wg.Done()
	return nil
}
