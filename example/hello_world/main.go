package main

import (
	"fmt"
	"log"
	"time"

	"github.com/minus5/nsqm"
	nsq "github.com/nsqio/go-nsq"
)

const (
	topic   = "hello_world"
	channel = "app"
)

func main() {
	// configuration
	cfgr := nsqm.Local()

	// create producer
	producer, err := nsqm.NewProducer(cfgr)
	if err != nil {
		log.Fatal(err)
	}

	// create consumer
	h := &handler{msgs: make(chan string)}
	consumer, err := nsqm.NewConsumer(cfgr, topic, channel, h)
	if err != nil {
		log.Fatal(err)
	}

	// send a message with producer
	msg := fmt.Sprintf("Hello Word at %s", time.Now())
	producer.Publish(topic, []byte(msg))

	// wait for consumer to receive a message
	log.Printf("received: %s\n", <-h.msgs)

	// cleanup
	producer.Stop()
	consumer.Stop()
}

type handler struct {
	msgs chan string
}

func (h *handler) HandleMessage(m *nsq.Message) error {
	fmt.Printf("received: %s\n", m.Body)
	h.msgs <- string(m.Body)
	return nil
}
