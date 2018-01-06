// The following directive is necessary to make the package coherent:

// +build ignore

// This program generates api for service.Service.
// It can be invoked by running:
// go generate
package main

import (
	"log"
	"reflect"

	"github.com/minus5/nsqm/example/rpc_with_consul_discovery/service"
	"github.com/minus5/nsqm/gen"
)

func main() {
	err := gen.Generate(gen.Config{
		ServiceType:      reflect.TypeOf(service.Service{}),
		NsqTopic:         "service.req",
		TransportTimeout: 2,
	})
	if err != nil {
		log.Fatal(err)
	}
}
