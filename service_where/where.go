package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"

	"github.com/xocasdashdash/envoy-test/common/request_id"
	"github.com/xocasdashdash/envoy-test/common/service_registry"

	"github.com/xocasdashdash/envoy-test/common/healthz"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
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
		r.Route("/where", func(r chi.Router) {
			r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
				fmt.Printf("New request about a location!\n")

				i := rand.Intn(len(places))
				place := places[i]
				status := http.StatusOK
				if place.Shadyness >= 90 {
					status = http.StatusBadGateway
				}
				rid := r.Context().Value(middleware.RequestIDKey).(string)
				if rid == "" {
					rid = "nope"
				}
				w.WriteHeader(status)

				if err := json.NewEncoder(w).Encode(place); err != nil {
					json.NewEncoder(w).Encode(struct{
						Error: err,
					})
					w.WriteHeader(http.StatusInternalServerError)
				}
			})

		})

	})
	r.HandleFunc("/_healthz", healthz.HealthCheck(ServiceName, 5))
	sr := service_registry.ServiceRegistry{
		Location: "consul:8500",
		Retries:  5,
	}

	if err := sr.Register("where", local_ip.GetLocalIP(), 3333); err != nil {
		log.Fatalf("Error registering service :(. Error: %+v", err)
	}
	http.Serve(l, r)

}
