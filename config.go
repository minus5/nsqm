package nsqm

import (
	"os"
	"time"

	"github.com/minus5/nsqm/discovery"
	nsq "github.com/nsqio/go-nsq"
)

// Config collects configuration parameters
type Config struct {
	NSQConfig           *nsq.Config
	NSQDAddress         string
	NSQLookupdAddresses []string
	Concurrency         int
	NodeName            string
	Logger              logger
	LogLevel            nsq.LogLevel
	dcy                 discoverer
}

type discoverer interface {
	NSQDAddress() (string, error)
	NSQLookupdAddresses() ([]string, error)
	Subscribe(discovery.Subscriber)
	NodeName() string
}

type logger interface {
	Output(calldepth int, s string) error
}

type nullLogger struct{}

func (nullLogger) Output(calldepth int, s string) error { return nil }

// Subscribe to nsqlookupd changes.
// Wehn location of nsqlookupd changes discovery will notify subscriber.
func (c *Config) Subscribe(subscriber discovery.Subscriber) {
	if c.dcy != nil {
		c.dcy.Subscribe(subscriber)
	}
}

func (c *Config) nsqConfig() *nsq.Config {
	if c.NSQConfig == nil {
		c.NSQConfig = nsq.NewConfig()
	}
	return c.NSQConfig
}

// Global defaults
var (
	MaxInFlight = 256
	Concurrency = 8
)

// Local returns Config for local nsqd.
// Usefull in development.
func Local() *Config {
	hostname, _ := os.Hostname()
	c := nsq.NewConfig()
	c.MaxInFlight = MaxInFlight
	return &Config{
		NSQDAddress:         "127.0.0.1:4150",
		NSQLookupdAddresses: nil,
		NSQConfig:           c,
		Concurrency:         Concurrency,
		NodeName:            hostname,
		Logger:              &nullLogger{},
	}
}

// WithDiscovery creates Config populated from discovery.x
func WithDiscovery(dcy discoverer) (*Config, error) {
	nsqd, err := dcy.NSQDAddress()
	if err != nil {
		return nil, err
	}
	lookups, err := dcy.NSQLookupdAddresses()
	if err != nil {
		return nil, err
	}
	c := nsq.NewConfig()
	c.LookupdPollInterval = 10 * time.Second
	c.MaxInFlight = MaxInFlight
	return &Config{
		NSQDAddress:         nsqd,
		NSQLookupdAddresses: lookups,
		NSQConfig:           c,
		Concurrency:         Concurrency,
		NodeName:            dcy.NodeName(),
		Logger:              &nullLogger{},
		dcy:                 dcy,
	}, nil
}
