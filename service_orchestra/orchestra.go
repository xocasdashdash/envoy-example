package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
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
	r.Use(middleware.RequestID)
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

	if err := registerService("hello", 3333); err != nil {
		log.Fatalf("Error registering service :(. Error: %+v", err)
	}
	http.Serve(l, r)

}
func registerService(service string, port int) error {
	config := api.DefaultConfig()
	config.Address = "consul:8500"
	client, err := api.NewClient(config)
	srvRegistrator := client.Agent()

	if err != nil {
		fmt.Printf("Encountered error connecting to consul on %s => %s\n", "consul", err)
		return err
	}
	rand.Seed(time.Now().UnixNano())
	sid := rand.Intn(65534)

	serviceID := service + "-" + strconv.Itoa(sid)

	consulService := api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    service,
		Tags:    []string{time.Now().Format("Jan 02 15:04:05.000 MST")},
		Port:    port,
		Address: GetLocalIP(),
		Checks:  api.AgentServiceChecks{},
	}

	return srvRegistrator.ServiceRegister(&consulService)

}
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}
