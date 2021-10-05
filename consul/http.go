package consul

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
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

func handlerGetServices(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logrus.Info(r.URL.Path[len(baseURL):])
		services := []*CatalogResponse{}
		fake := &CatalogResponse{ServiceName: "toto", Address: "xxx", ServicePort: 2343}
		services = append(services, fake)
		out, err := json.Marshal(services)
		if err != nil {
			w.WriteHeader(500)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(fmt.Sprint("Error: ", err.Error())))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	}

}

func HandleServices() {
	baseURL := "/v1/catalog/service/"
	http.HandleFunc(baseURL, handlerGetServices(baseURL))
}
