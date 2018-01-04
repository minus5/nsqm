package nsqm

import nsq "github.com/nsqio/go-nsq"

const (
	defaultConcurrency = 256
)

type Configurator interface {
	NSQDAddress() string
	NSQLookupdAddresses() []string
	Config() *nsq.Config
	Concurrency() int
	Output(calldepth int, s string) error
	SetLookupdDiscovery(lookupdDiscovery)
}

type lookupdDiscovery interface {
	DisconnectFromNSQLookupd(addr string) error
	ConnectToNSQLookupd(addr string) error
}

func NewProducer(cfgr Configurator) (*nsq.Producer, error) {
	producer, err := nsq.NewProducer(cfgr.NSQDAddress(), cfgr.Config())
	if err != nil {
		return nil, err
	}
	producer.SetLogger(cfgr, nsq.LogLevelDebug)
	return producer, nil
}

func NewConsumer(cfgr Configurator, topic, channel string, handler nsq.Handler) (*nsq.Consumer, error) {
	consumer, err := nsq.NewConsumer(topic, channel, cfgr.Config())
	if err != nil {
		return nil, err
	}
	consumer.SetLogger(cfgr, nsq.LogLevelDebug)
	consumer.AddConcurrentHandlers(handler, cfgr.Concurrency())
	if addrs := cfgr.NSQLookupdAddresses(); addrs != nil {
		if err := consumer.ConnectToNSQLookupds(addrs); err != nil {
			return nil, err
		}
	} else {
		if err := consumer.ConnectToNSQD(cfgr.NSQDAddress()); err != nil {
			return nil, err
		}
	}
	cfgr.SetLookupdDiscovery(consumer)
	return consumer, nil
}

func Local() Configurator {
	return &localConfigurator{}
}

type localConfigurator struct{}

func (c *localConfigurator) NSQDAddress() string {
	return "127.0.0.1:4150"
}

func (c *localConfigurator) NSQLookupdAddresses() []string {
	return nil
}

func (c *localConfigurator) Config() *nsq.Config {
	return nsq.NewConfig()
}

func (c *localConfigurator) Output(calldepth int, s string) error {
	return nil
}

func (c *localConfigurator) Concurrency() int {
	return defaultConcurrency
}

func (c *localConfigurator) SetLookupdDiscovery(ld lookupdDiscovery) {}

type discovery interface {
	NSQDAddress() (string, error)
	NSQLookupdAddresses() ([]string, error)
}

func WithDiscovery(dcy discovery) Configurator {
	return &discoveryConfigurator{dcy: dcy}
}

type discoveryConfigurator struct {
	dcy discovery
}

func (c *discoveryConfigurator) NSQDAddress() string {
	addr, _ := c.dcy.NSQDAddress()
	return addr
}

func (c *discoveryConfigurator) NSQLookupdAddresses() []string {
	addrs, _ := c.dcy.NSQLookupdAddresses()
	return addrs
}

func (c *discoveryConfigurator) Config() *nsq.Config {
	return nsq.NewConfig()
}

func (c *discoveryConfigurator) Output(calldepth int, s string) error {
	return nil
}

func (c *discoveryConfigurator) Concurrency() int {
	return defaultConcurrency
}

func (c *discoveryConfigurator) SetLookupdDiscovery(ld lookupdDiscovery) {

}
