# Simple Service registry based on pub/sub pattern

This first version service registry use nats (https://nats.io/) to implement a service registry
You just need to start nats before using this service registry

The project is still under development and is not ready for production. 

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
    //connect to nats server
	c, err := nats.Connect("localhost:4222")
	if err != nil {
		log.Fatal("Could not connect to nats: ", err)
	}
    //Create registry instance
	reg, err := registry.Connect(registry.Nats(c))
	if err != nil {
		log.Fatal("Could not open registry session", err)
	}
	//Register
    unregister, err := reg.Register(registry.Service{Name: "httptest", URL: "http://localhost:8083/test"})
    if err != nil {
        log.Fatal("Could not register the service ", err)
    }
    //call the unregister func when the service is shuting done
    defer unregister()
    // Start your service ...
    
```

On the client side :
```golang
    //connect to nats server
	c, err := nats.Connect("localhost:4222")
	if err != nil {
		log.Fatal("Could not connect to nats: ", err)
	}
    //Create registry instance
	reg, err := registry.Connect(registry.Nats(c))
	if err != nil {
		log.Fatal("Could not open registry session", err)
    }
    //registry service register to nats all event for the service httptest
    //This is optional, registry
    err = registry.Observe("httptest")
    if err != nil {
        log.Error("Could not observe service ", err)
        return
    }

    ctx, fct := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer fct()
	services, err := r.GetServices(ctx, "httptest")
	if err != nil {
        log.Error("Could not get services ", err)
        return
    }
    //Do something with services
```