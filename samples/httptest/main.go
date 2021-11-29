package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/eddieraa/registry"
	pb "github.com/eddieraa/registry/nats"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

func main() {

	var serviceName string
	flag.StringVar(&serviceName, "service-name", "httptest", "http service name")
	var natsURL string
	flag.StringVar(&natsURL, "nats-url", "localhost:4222", "NATS server URL ")
	var close bool
	flag.BoolVar(&close, "close", true, "do not unregister")
	var warning bool
	flag.BoolVar(&warning, "status-warning", false, "Set this service in warning status")

	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)

	addr, err := registry.FindFreeLocalAddress(10000, 10020)
	if err != nil {
		panic(fmt.Sprint("Could not get free port ", err))
	}
	println("Start http server on address (", addr, ")")
	conn, err := nats.Connect(natsURL)
	if err != nil {
		panic(fmt.Sprint("Could not connect to nats (", natsURL, "): ", err))
	}
	r, err := registry.SetDefault(pb.Nats(conn), registry.WithLoglevel(logrus.DebugLevel))
	if err != nil {
		panic(fmt.Sprint("Could not create registry ", err))
	}
	service := registry.Service{Name: serviceName, Address: addr}
	unregister, err := r.Register(service)
	if warning {
		r.SetServiceStatus(service, registry.Warning)
	}

	//Intercep CTRL-C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	handler := http.NewServeMux()
	s := &http.Server{Addr: addr, Handler: handler}
	count := 0
	handler.HandleFunc("/httptest", func(out http.ResponseWriter, req *http.Request) {

		var buf bytes.Buffer
		buf.WriteString(fmt.Sprint("Hello from ", addr, " URL: ", req.URL, " RequestURI: ", req.RequestURI+" host("+req.URL.Host+") path("+req.URL.Path+") rawPath ("+req.URL.RawPath+")\n"))
		//buf.WriteString(fmt.Sprint("Header: ", req.Header))
		// Loop through headers
		for name, headers := range req.Header {
			name = strings.ToLower(name)
			for _, h := range headers {
				buf.WriteString(fmt.Sprintf("%v: %v\n", name, h))
			}
		}

		out.Write(buf.Bytes())
		if count%100 == 0 {
			println()
			logrus.Print(count)
		} else {
			print(".")
		}
		count++
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
