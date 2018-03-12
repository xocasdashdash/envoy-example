package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/hashicorp/consul/api"
)

const version = 0
const ServiceName = "where_service"

type Place struct {
	Location  string
	Shadyness int
}

var places []Place

func init() {
	places = []Place{
		Place{"Home", 0},
		Place{"Supermarket", 0},
		Place{"Work", 0},
		Place{"Shop", 0},
		Place{"Gym", 0},
		Place{"Workshop", 0},
		Place{"Corner", 90},
		Place{"Pool", 10},
	}

}

var reqid uint64

func RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		for k, v := range r.Header {
			fmt.Printf("Header field %q, Value %q\n", k, v)
		}
		rid := r.Header.Get("X-Request-Id")
		if rid == "" {
			myid := atomic.AddUint64(&reqid, 1)
			rid = fmt.Sprintf("%d-%d-%d-%d", myid)
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, middleware.RequestIDKey, fmt.Sprintf("%s", rid))
		w.Header().Add("X-Request-Id", rid)
		fmt.Printf("Serving request with id: %s\n", rid)
		next.ServeHTTP(w, r.WithContext(ctx))
		fmt.Printf("Served request with id: %s\n", rid)
	}
	return http.HandlerFunc(fn)
}
func (p Place) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
func main() {
	r := chi.NewRouter()
	r.Use(RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.StripSlashes)
	l, err := net.Listen("tcp", ":3333")
	if err != nil {
		log.Fatal(err)
	}
	r.Route(fmt.Sprintf("/v%d/", version), func(r chi.Router) {

		r.Route("/where", func(r chi.Router) {
			r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
				fmt.Printf("New request about a location!\n")

				i := rand.Intn(len(places))
				place := places[i]
				status := http.StatusOK
				if place.Shadyness >= 90 {
					status = http.StatusBadGateway
				}
				if status >= 300 {
					w.Header().Add("Alive", "false")
				} else {
					w.Header().Add("Alive", "true")

				}
				rid := r.Context().Value(middleware.RequestIDKey).(string)
				if rid == "" {
					rid = "nope"
				}
				w.WriteHeader(status)
				if err := render.Render(w, r, place); err != nil {
					render.Render(w, r, ErrRender(err))
					return
				}
			})

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

	if err := registerService("where", 3333); err != nil {
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
