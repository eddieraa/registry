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

func waitNatsAndRegister(natsUrl string, address string) {
	for {
		conn, err := nats.Connect(natsUrl, nats.Timeout(time.Second*5))
		if err != nil {
			logrus.Info("Could not connect to nats (", natsUrl, "): ", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if _, err = registry.SetDefault(pb.Nats(conn), registry.WithLoglevel(logrus.DebugLevel)); err != nil {
			logrus.Error(fmt.Sprint("Could not create registry ", err))
			break
		}
		if _, err = registry.Register(registry.Service{Name: "http-test2", Address: address}); err != nil {
			logrus.Error("could not register service ", err)
		} else {
			logrus.Info("register ok")
		}
		break
	}
}

func main() {

	var natsURL string
	flag.StringVar(&natsURL, "nats-url", "localhost:4222", "NATS server URL ")
	var addr string
	flag.StringVar(&addr, "addr", ":8181", "address to listen")

	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)

	go waitNatsAndRegister(natsURL, addr)
	defer registry.Close()
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

		s.Shutdown(ctx)
	}()

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprint("Could not start http server on address ", addr, " err: ", err))
	}
}
