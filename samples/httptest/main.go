package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eddieraa/registry"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

func localfreeaddr() string {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(fmt.Sprint("Could not open new port: ", err))
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(fmt.Sprint("Error listen addr ", addr, " err: ", err))
	}
	port := l.Addr().(*net.TCPAddr).Port

	defer l.Close()
	return fmt.Sprint("127.0.0.1:", port)
}

func main() {

	var serviceName string
	flag.StringVar(&serviceName, "service-name", "httptest", "http service name")
	var natsURL string
	flag.StringVar(&natsURL, "nats-url", "localhost:4222", "NATS server URL ")
	var close bool
	flag.BoolVar(&close, "close", true, "do not unregister")

	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)

	addr := localfreeaddr()
	println("Start http server on address ", addr)
	conn, err := nats.Connect(natsURL)
	if err != nil {
		panic(fmt.Sprint("Could not connect to nats (", natsURL, "): ", err))
	}
	r, err := registry.Connect(registry.Nats(conn))
	if err != nil {
		panic(fmt.Sprint("Could not create registry ", err))
	}
	unregister, err := r.Register(registry.Service{Name: serviceName, Address: addr})

	//Intercep CTRL-C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	s := &http.Server{Addr: addr, Handler: &http.ServeMux{}}
	s.Handler = http.HandlerFunc(func(out http.ResponseWriter, req *http.Request) {
		out.Write([]byte(fmt.Sprint("Hello from ", addr, "\n")))
	})
	go func() {
		<-sigs
		fmt.Printf("Stop and unregister\n")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if close {
			unregister()
			r.Close()
		}

		s.Shutdown(ctx)
	}()

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprint("Could not start http server on address ", addr, " err: ", err))
	}
}
