package consul

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/minus5/nsqm/discovery"
)

const (
	nsqLookupdHTTPServiceName      = "nsqlookupd-http"
	nsqLookupdHTTPServiceNameByTag = "nsqlookupd"
	nsqLookupdHTTPServiceTag       = "http"
	nsqdTCPServiceName             = "nsqd-tcp"
	nsqdTCPServiceNameByTag        = "nsqd"
	nsqdTCPServiceTag              = "tcp"
)

// Local creates discovery with local consul addres.
func Local() (*dcy, error) {
	return New("127.0.0.1:8500")
}

// New creates discovery using provided consul address.
func New(addr string) (*dcy, error) {
	cfg := api.DefaultConfig()
	cfg.Address = addr
	cli, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &dcy{cli: cli, addr: addr}, nil
}

type dcy struct {
	addr         string
	cli          *api.Client
	lookupdAddrs []string
	subscribers  []discovery.Subscriber
	sync.Mutex
	monitorOnce sync.Once
}

type service struct {
	name string
	tag  string
}

func (d *dcy) NSQDAddress() (string, error) {
	s := []service{{nsqdTCPServiceName, ""}, {nsqdTCPServiceNameByTag, nsqdTCPServiceTag}}
	addr, err := d.agentService(s)
	if err != nil {
		return "", err
	}
	if addr != "" {
		return addr, err
	}

	ses, err := d.service(nsqdTCPServiceName, "")
	if err != nil {
		return "", err
	}
	addrs := parseServiceEntries(ses)
	if len(addrs) != 0 {
		return addrs[0], nil
	}
	ses, err = d.service(nsqdTCPServiceNameByTag, nsqdTCPServiceTag)
	if err != nil {
		return "", err
	}
	addrs = parseServiceEntries(ses)
	if len(addrs) == 0 {
		return "", fmt.Errorf("not found")
	}
	return addrs[0], nil
}

func (d *dcy) agentService(services []service) (string, error) {
	svcs, err := d.cli.Agent().Services()
	if err != nil {
		return "", err
	}
	for _, svc := range svcs {
		for _, s := range services {
			if svc.Service == s.name && contains(svc.Tags, s.tag) {
				addr := svc.Address
				if addr == "" {
					addr = d.consulAddr()
				}
				return fmt.Sprintf("%s:%d", addr, svc.Port), nil
			}
		}
	}
	return "", nil
}

func (d *dcy) consulAddr() string {
	return strings.Split(d.addr, ":")[0]
}

func (d *dcy) NSQLookupdAddresses() ([]string, error) {
	d.Lock()
	defer d.Unlock()
	if d.lookupdAddrs != nil {
		return d.lookupdAddrs, nil
	}
	ses, err := d.service(nsqLookupdHTTPServiceName, "")
	if err != nil {
		return nil, err
	}
	lookupdAddrs := parseServiceEntries(ses)
	if len(lookupdAddrs) != 0 {
		d.lookupdAddrs = lookupdAddrs
		return d.lookupdAddrs, nil
	}

	ses, err = d.service(nsqLookupdHTTPServiceNameByTag, nsqLookupdHTTPServiceTag)
	if err != nil {
		return nil, err
	}
	d.lookupdAddrs = parseServiceEntries(ses)
	return d.lookupdAddrs, nil
}

func (d *dcy) service(name, tag string) ([]*api.ServiceEntry, error) {
	ses, _, err := d.cli.Health().Service(name, tag, true, nil)
	return ses, err
}

func parseServiceEntries(ses []*api.ServiceEntry) []string {
	var addrs []string
	for _, se := range ses {
		addr := se.Service.Address
		if addr == "" {
			addr = se.Node.Address
		}
		addrs = append(addrs, fmt.Sprintf("%s:%d", addr, se.Service.Port))
	}
	return addrs
}

func (d *dcy) monitor() {
	var wi uint64
	for {
		qo := &api.QueryOptions{
			WaitIndex:         wi,
			WaitTime:          time.Minute,
			AllowStale:        true,
			RequireConsistent: false,
		}
		ses, qm, err := d.cli.Health().Service(nsqLookupdHTTPServiceName, "", true, qo)
		if err != nil {
			//log.Printf("error: %s", err) // TODO
			time.Sleep(time.Second)
			continue
		}
		if len(ses) == 0 {
			ses, qm, err = d.cli.Health().Service(nsqLookupdHTTPServiceNameByTag, nsqLookupdHTTPServiceTag, true, qo)
			if err != nil {
				//log.Printf("error: %s", err) // TODO
				time.Sleep(time.Second)
				continue
			}
		}
		addrs := parseServiceEntries(ses)
		d.updateLookups(addrs)
		wi = qm.LastIndex
	}
}

func (d *dcy) Subscribe(s discovery.Subscriber) {
	d.Lock()
	defer d.Unlock()
	d.subscribers = append(d.subscribers, s)
	d.monitorOnce.Do(func() {
		go d.monitor()
	})
}

func (d *dcy) updateLookups(addrs []string) {
	d.Lock()
	defer d.Unlock()
	contains := func(addrs []string, addr string) bool {
		for _, a := range addrs {
			if a == addr {
				return true
			}
		}
		return false
	}
	changed := false
	for _, subscriber := range d.subscribers {
		for _, addr := range addrs {
			// add newly discovered lookupd
			if !contains(d.lookupdAddrs, addr) {
				changed = true
				//fmt.Println("ConnectToNSQLookupd", addr) // TODO
				if err := subscriber.ConnectToNSQLookupd(addr); err != nil {
					// TODO logging
					//log.Printf("error: %s", err) // TODO
				}
			}
		}
		for _, addr := range d.lookupdAddrs {
			// remove lookupd which don't exists any more
			if !contains(addrs, addr) {
				changed = true
				//fmt.Println("DisconnectFromNSQLookupd", addr) // TODO
				if err := subscriber.DisconnectFromNSQLookupd(addr); err != nil {
					// TODO logging
					//log.Printf("error: %s", err) // TODO
				}
			}
		}
	}
	if changed {
		//fmt.Printf("updating lookupds to %v\n", addrs)
		d.lookupdAddrs = addrs
	}
}

// testing register example:
// curl -s -X PUT -d '{"Node":"app1","Address":"10.0.66.157","Service":{"Service":"nsqlookupd-http","Port":10901}}' http://127.0.0.1:8500/v1/catalog/register

func (d *dcy) NodeName() string {
	s, err := d.cli.Agent().Self()
	if err != nil {
		return ""
	}
	cfg := s["Config"]
	return cfg["NodeName"].(string)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
