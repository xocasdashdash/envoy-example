package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

func extractName(r *http.Request) string {
	name := chi.URLParam(r, "name")
	if name == "" {
		name = "Anonymous"
	}
	return name
}

const version = 0
const ServiceName = "hello_service"

type Greeting struct {
	Name    string `json:name`
	Time    string `json:time`
	Message string `json:message`
}

func (s Greeting) Render(w http.ResponseWriter, r *http.Request) error {

	return nil
}

var reqid uint64

func RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		for k, v := range r.Header {
			fmt.Fprintf(w, "Header field %q, Value %q\n", k, v)
		}
		rid := r.Header.Get("X-Request-Id")
		if rid == "" {
			myid := atomic.AddUint64(&reqid, 1)
			rid = fmt.Sprintf("%d-%d-%d-%d", myid)
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, middleware.RequestIDKey, fmt.Sprintf("%s", rid))
		w.Header().Add("X-Request-Id", rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
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
	closeChan := make(chan bool)
	r.Route(fmt.Sprintf("/v%d/", version), func(r chi.Router) {
		r.Get("/hello/{name}", func(w http.ResponseWriter, req *http.Request) {

			name := extractName(req)
			greet := Greeting{
				Name:    name,
				Time:    "1234",
				Message: "Solo se habla español!Adiós!",
			}
			w.WriteHeader(http.StatusBadRequest)

			closeChan <- true
			if err := render.Render(w, req, greet); err != nil {
				render.Render(w, req, ErrRender(err))
				return
			}

		})
		r.Get("/hola/{name}", func(w http.ResponseWriter, req *http.Request) {
			name := extractName(req)
			greet := Greeting{
				Name:    name,
				Time:    "1234",
				Message: "Hola my friend!",
			}
			if err := render.Render(w, req, greet); err != nil {
				render.Render(w, req, ErrRender(err))
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
	srv := &http.Server{Addr: l.Addr().String(), Handler: r}
	ctx := context.Background()
	go func() {
		if err := registerService("hello", 3333, 5); err != nil {
			log.Printf("Error registering service :(. Error: %+v", err)
			srv.Shutdown(ctx)
		}
	}()
	go func() {
		if err := srv.Serve(l); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Fatalf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()
	log.Printf("Waiting for close signal")
	<-closeChan
	log.Printf("Received close signal")
	log.Print(srv.Shutdown(ctx))
}
func registerService(service string, port int, retries int) error {
	config := api.DefaultConfig()
	retriesLeft := retries
	var srvRegistrator *api.Agent
	for {
		log.Printf("Trying to connect to consul...")
		config.Address = "consul:8500"
		client, _ := api.NewClient(config)
		srvRegistrator = client.Agent()
		//Chec connection to consul
		_, err := srvRegistrator.Self()

		if err != nil && retriesLeft <= 0 {
			log.Printf("Encountered error connecting to consul on %s => %s\n", "consul", err)
			return err
		} else if err != nil {
			waitTime := time.Duration(math.Exp2(float64(retries-retriesLeft))) * time.Second
			log.Printf("Encountered error connecting to consul on %s => %s\n", "consul", err)
			log.Printf("Waiting %.0f secs until next connection attempt", waitTime.Seconds())
			log.Printf("Retries left: %d", retriesLeft)
			retriesLeft = retriesLeft - 1
			time.Sleep(waitTime)
		} else {
			log.Printf("Connection successful!")
			break
		}

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
