package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/xocasdashdash/envoy-example/common/local_ip"
	"github.com/xocasdashdash/envoy-example/common/request_id"
	"github.com/xocasdashdash/envoy-example/common/service_registry"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const version = 0
const ServiceName = "orchestra_service"

var numberOfReqs uint64
var totalRequests uint64
var proxyName string

func init() {
	numberOfReqs = uint64(rand.Int63n(200) + 100)
	totalRequests = 0
	proxyName = os.Getenv("PROXY_URL")
}

var ReqIDHeader string = "X-Request-Id"
var serviceClient = &http.Client{Timeout: 10 * time.Second}

func getJson(service string, path string, headers http.Header, target interface{}) error {
	urlString := fmt.Sprintf("http://%s/%s/%s", proxyName, service, path)
	log.Printf("Sending request to url:%s", urlString)
	req, err := http.NewRequest("GET", urlString, nil)
	if rid := headers.Get(ReqIDHeader); rid != "" {
		req.Header.Add(ReqIDHeader, rid)
	}
	resp, err := serviceClient.Do(req)
	if err != nil {
		return err
	}
	buff, _ := ioutil.ReadAll(resp.Body)
	log.Printf("Message received: %s. Status: %s", string(buff), resp.Status)
	defer resp.Body.Close()

	return nil
}
func main() {
	r := chi.NewRouter()
	r.Use(request_id.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.StripSlashes)

	l, err := net.Listen("tcp", ":3333")
	if err != nil {
		log.Fatal(err)
	}
	r.Route(fmt.Sprintf("/v%d/", version), func(r chi.Router) {
		r.Get("/orchestra", func(w http.ResponseWriter, r *http.Request) {
			totalRequests = atomic.AddUint64(&totalRequests, 1)
			getJson("where", "/", r.Header, nil)
			getJson("who", "/", r.Header, nil)

		})

	})
	r.HandleFunc("/_healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("x-envoy-upstream-healthchecked-cluster", ServiceName)
		name, _ := os.Hostname()
		w.Header().Add("x-upstream-server", name)
		if totalRequests >= numberOfReqs {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

	})
	sr := service_registry.ServiceRegistry{
		Location: "consul:8500",
		Retries:  5,
	}

	if err := sr.Register("orchestra", local_ip.GetLocalIP(), 3333); err != nil {
		log.Fatalf("Error registering service :(. Error: %+v", err)
	}
	http.Serve(l, r)

}
