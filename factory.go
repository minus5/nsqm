package nsqm

import nsq "github.com/nsqio/go-nsq"

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
