package healthz

import (
	"math/rand"
	"net/http"
	"os"
)

func HealthCheck(ServiceName string, KOSeed int) http.HandlerFunc {
	hostName, _ := os.Hostname()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("x-envoy-upstream-healthchecked-cluster", ServiceName)
		w.Header().Add("x-upstream-host", hostName)

		if (rand.Intn(KOSeed) + 1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("ko"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}
	}

}
