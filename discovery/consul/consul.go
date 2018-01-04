package consul

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

const (
	nsqLookupdHTTPServiceName = "nsqlookupd-http"
	nsqdTCPServiceName        = "nsqd-tcp"
)

func New(addr string) (*discovery, error) {
	cfg := api.DefaultConfig()
	cfg.Address = addr
	cli, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &discovery{cli: cli, addr: addr}, nil
}

type discovery struct {
	addr string
	cli  *api.Client
}

func (d *discovery) NSQDAddress() (string, error) {
	addr, err := d.agentService(nsqdTCPServiceName)
	if err != nil {
		return "", err
	}
	if addr != "" {
		return addr, err
	}

	ses, err := d.service(nsqdTCPServiceName)
	if err != nil {
		return "", err
	}
	addrs := parseServiceEntries(ses)
	if len(addrs) == 0 {
		return "", fmt.Errorf("not found")
	}
	return addrs[0], nil
}

func (d *discovery) agentService(name string) (string, error) {
	svcs, err := d.cli.Agent().Services()
	if err != nil {
		return "", err
	}
	for _, svc := range svcs {
		if svc.Service == name {
			addr := svc.Address
			if addr == "" {
				addr = d.addr
			}
			return fmt.Sprintf("%s:%d", addr, svc.Port), nil
		}
	}
	return "", nil
}

func (d *discovery) NSQLookupdAddresses() ([]string, error) {
	ses, err := d.service(nsqLookupdHTTPServiceName)
	if err != nil {
		return nil, err
	}
	return parseServiceEntries(ses), nil
}

func (d *discovery) service(name string) ([]*api.ServiceEntry, error) {
	ses, _, err := d.cli.Health().Service(name, "", true, nil)
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
