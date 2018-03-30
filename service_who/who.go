package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/xocasdashdash/envoy-test/common/healthz"
	"github.com/xocasdashdash/envoy-test/common/local_ip"
	"github.com/xocasdashdash/envoy-test/common/request_id"
	"github.com/xocasdashdash/envoy-test/common/service_registry"
)

const version = 0
const ServiceName = "who_service"

type Person struct {
	Name       string
	Occupation string
	Age        string
}

var people []Person
var hostName string

func init() {
	people = []Person{
		Person{
			Name:       "Napoleon",
			Occupation: "Not Russian leader",
			Age:        "n.d.",
		},
		Person{
			Name:       "George Washington",
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
	hostName, _ = os.Hostname()
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
		r.Get("/who", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(people[rand.Intn(len(people))])
		})

	})

	r.HandleFunc("/_healthz", healthz.HealthCheck("hello", 5))

	sr := service_registry.ServiceRegistry{
		Location: "consul:8500",
		Retries:  5,
	}
	if err := sr.Register("hello", local_ip.GetLocalIP(), 3333); err != nil {
		log.Fatalf("Error registering service :(. Error: %+v", err)
	}

	http.Serve(l, r)
}
