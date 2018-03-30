package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/xocasdashdash/envoy-example/common/healthz"
	"github.com/xocasdashdash/envoy-example/common/request_id"
	"github.com/xocasdashdash/envoy-example/common/service_registry"

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
			if err := json.NewEncoder(w).Encode(greet); err != nil {
				json.NewEncoder(w).Encode(struct{
					Error: err,
				})
				w.WriteHeader(http.StatusInternalServerError)
			}

		})
		r.Get("/hola/{name}", func(w http.ResponseWriter, req *http.Request) {
			name := extractName(req)
			greet := Greeting{
				Name:    name,
				Time:    "1234",
				Message: "Hola my friend!",
			}

			if err := json.NewEncoder(w).Encode(greet); err != nil {
				json.NewEncoder(w).Encode(struct{
					Error: err,
				})
				w.WriteHeader(http.StatusInternalServerError)
			}
		})
	})
	r.HandleFunc("/_healthz", healthz.HealthCheck(ServiceName, 5))
	srv := &http.Server{Addr: l.Addr().String(), Handler: r}
	ctx := context.Background()
	go func() {
		sr := service_registry.ServiceRegistry{
			Location: "consul:8500",
			Retries:  5,
		}

		if err := sr.Register("where", local_ip.GetLocalIp(), 3333); err != nil {
			log.Fatalf("Error registering service :(. Error: %+v", err)
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
