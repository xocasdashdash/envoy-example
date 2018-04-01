package request_id

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/go-chi/chi/middleware"
)

var reqid uint64

func RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		for k, v := range r.Header {
			log.Printf("Header field %q, Value %q\n", k, v)
		}
		rid := r.Header.Get("X-Request-Id")
		if rid == "" {
			myid := atomic.AddUint64(&reqid, 1)
			rid = fmt.Sprintf("%d", myid)
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, middleware.RequestIDKey, fmt.Sprintf("%s", rid))
		w.Header().Add("X-Request-Id", rid)
		log.Printf("Serving request with id: %s\n", rid)
		next.ServeHTTP(w, r.WithContext(ctx))
		log.Printf("Served request with id: %s\n", rid)
	}
	return http.HandlerFunc(fn)
}
