package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

const version = 0
const ServiceName = "hello_service"

type Person struct {
	Name       string
	Occupation string
	Age        string
}

var people []Person

func init() {
	people = []Person{
		Person{
			Name:       "Napoleon",
			Occupation: "Not Russia",
			Age:        "n.d.",
		},
		Person{
			Name:       "Washington",
			Occupation: "Freedom Fighter",
			Age:        "42",
		}, Person{
			Name:       "San Martín",
			Occupation: "Libertador",
			Age:        "60",
		}, Person{
			Name:       "Horatio Nelson",
			Occupation: "Admiral.",
			Age:        "35",
		},
	}
}
func (p Person) Render(w http.ResponseWriter, r *http.Request) error {

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
		r.Get("/who", func(w http.ResponseWriter, r *http.Request) {
			dice := rand.Intn(len(people))
			p := people[dice]
			if err := render.Render(w, r, p); err != nil {
				render.Render(w, r, ErrRender(err))
				return
			}
		})

	})
	r.HandleFunc("/_healthz", func(w http.ResponseWriter, r *http.Request) {
		dice := rand.Intn(5) + 1
		w.Header().Add("x-envoy-upstream-healthchecked-cluster", ServiceName)

		if dice == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("ko"))
		} else {
			w.Write([]byte("ok"))
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
