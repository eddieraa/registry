# Simple Service registry based on pub/sub pattern

[![License Apache 2](https://img.shields.io/badge/License-Apache2-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)

This first version service registry use nats (https://nats.io/) or reids (https://redis.io/) to implement a service registry
You just need to start nats before using this service registry

The project is still under development and is not ready for production. 

A java [implentation](https://github.com/eddieraa/registry-java) exist

## Simple to use: 
API is simple to use, simple to understand. No required external installation.
No configuration file is required.
- Connect: `registry.Connect(registry.Nats(c))` init the service both on the client/server side and return registry service instance.
- Register: `reg.Register(registry.Service{Name: "httptest", URL: "http://localhost:8083/test"})` register you service on the service side.
- GetService: `reg.GetService("myservice")` on the client side
- Unregister: `reg.Unregister("myservice")` on the service side
- Close : `reg.Close()` on the server side 'Close' function deregisters all registered services. On the client and service side, unsubsribe to subscriptions

## Safe

Can be used in concurrent context. Singleton design pattern is used.
The first called to `registry.Connect(registry.Nats(c), registry.Timeout(100*time.MilliSecond))` create instance of registry service. 
And then you can get an instance of registry service without argument `reg, _ := registry.Connect()` 

## Fast
This "service registry" uses a local cache as well as the pub / sub pattern. The registration of a new service is in real time.
When a new service registers, all clients are notified in real time and update their cache.

## Light
The project uses very few external dependencies. There is no service to install. Just your favorite pub/sub implementation (NATS)

## Multi-tenant
Multi-tenant support is planned. The tool is based on pub / sub, each message can be pre-fixed by a text of your choice.

## Exemple

on the server side :
```golang
    package main

import (
	"log"
	"net/http"

	"github.com/eddieraa/registry"
	rnats "github.com/eddieraa/registry/nats"
	"github.com/nats-io/nats.go"
)

func main() {
	//connect to nats server
	conn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Could not connect to nats ", err)
	}

	//create registry default instance
	reg, err := rnats.SetDefault(conn)
	if err != nil {
		log.Fatal("Could not create registry instance ", err)
	}

	addr, _ := registry.FindFreeLocalAddress(10000, 10010)
	//register "myservice"
	fnuregister, err := reg.Register(registry.Service{Name: "myservice", Address: addr})
	if err != nil {
		log.Fatal("Could not register service ", err)
	}

	server := http.NewServeMux()
	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
	if err = http.ListenAndServe(addr, server); err != nil {
		log.Fatal("could not create server ", err)
	}

	//Unregister
	fnuregister()
	reg.Close()
	conn.Close()

}
    
```

On the client side :
```golang
package main

import (
	"log"

	rnats "github.com/eddieraa/registry/nats"
	"github.com/nats-io/nats.go"
)

func main() {
	//connect to nats server
	conn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Could not connect to nats ", err)
	}

	//create registry default instance
	reg, err := rnats.SetDefault(conn)
	if err != nil {
		log.Fatal("Could not create registry instance ", err)
	}

	//lookup for "myservice"
	s, err := reg.GetService("myservice")
	if err != nil {
		log.Fatal("Could not get service")
	}

	//do something with your service
	log.Print("Find service with address ", s.Address)

	reg.Close()
	conn.Close()

}

```
