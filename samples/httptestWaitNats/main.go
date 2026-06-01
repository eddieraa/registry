package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/eddieraa/registry"
	nregistry "github.com/eddieraa/registry/nats"
)

func main() {

	var natsURL string
	flag.StringVar(&natsURL, "nats-url", "localhost:4222", "NATS server URL ")
	var addr string
	flag.StringVar(&addr, "addr", ":8181", "address to listen")

	flag.Parse()
	var log = slog.Default()
	log.Debug("start service")

	go func() {

		//init register without existing nats connection
		nregistry.SetDefault(nil, nregistry.WithNatsUrl(natsURL))
		//execute Register in go routine
		//Register will block until nats connection is OK
		registry.Register(registry.Service{Address: addr, Name: "sample2", KV: map[string]string{"node": "1"}})
		log.Debug("service registered")
	}()
	defer registry.Close()
	log.Debug("registgry created")

	handler := http.NewServeMux()
	s := &http.Server{Addr: addr, Handler: handler}
	handler.HandleFunc("/httptest", func(out http.ResponseWriter, req *http.Request) {
		out.Write([]byte("httptest OK"))
	})

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprint("Could not start http server on address ", addr, " err: ", err))
	}
}
