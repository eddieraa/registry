package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eddieraa/registry"
)

const (
	REGISTRY_NAME = "registry-http-service"
)

type CatalogResponse struct {
	ID             string
	Node           string
	Address        string
	Datacenter     string
	NodeMeta       map[string]string
	ServiceAddress string
	ServiceID      string
	ServiceName    string
	ServicePort    int
	ServiceMeta    map[string]string
}

func handlerGetServices(reg registry.Registry, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := r.URL.Query().Has("indent")
		serviceName := r.URL.Path[len(baseURL):]

		var serviceNames []string
		if serviceName == "*" {
			serviceNames = reg.GetObservedServiceNames()
		} else {
			serviceNames = []string{serviceName}
		}

		resp := []*CatalogResponse{}
		for _, sn := range serviceNames {
			services, err := reg.GetServices(sn)
			if err != nil {
				sendError(w, err)
				return
			}

			for _, s := range services {
				resp = append(resp, &CatalogResponse{
					ServiceName: s.Name,
					Address:     s.Host,
					ServicePort: registry.Port(s),
				})
			}
		}
		var out []byte
		var err error
		if format {
			out, err = json.MarshalIndent(resp, "", "  ")
		} else {
			out, err = json.Marshal(resp)
		}

		if err != nil {
			sendError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	}

}

func sendError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprint("Error: ", err.Error())))
}

func HandleServices(r registry.Registry) {
	baseURL := "/v1/catalog/service/"
	http.Handle(baseURL, handlerGetServices(r, baseURL))
}
