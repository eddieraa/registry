# Service registry based on NATS protocol

This service registry use nats (https://nats.io/) for implement a service registry
You just need to start nats before using this service registry

on the server side :
```golang
    //connect to nats server
	c, err := nats.Connect("localhost:4222")
	if err != nil {
		log.Fatal("Could not connect to nats: ", err)
	}
    //Create registry instance
	reg, err := registry.Connect(c)
	if err != nil {
		log.Fatal("Could not open registry session", err)
	}
	host, _ := os.Hostname()
    unregister, err := reg.Register(registry.Service{Name: "httptest", Address: "localhost:8083", Network: "tcp", Host: host})
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
	reg, err := registry.Connect(c)
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