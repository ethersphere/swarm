package http

import (
	"net/http"

	"github.com/ethereum/go-ethereum/swarm/api"
)

//   http.Handle("/", Adapt(indexHandler, AddHeader("Server", "Mine"),
//                                      CheckAuth(providers),
//                                      CopyMgoSession(db),
//                                      Notify(logger),
// 								   )

type RequestState struct {
	ReqId string
	Uri   *api.URI
	Error error
	Code  uint
}

func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range adapters {
		h = adapter(h)
	}
	return h
}

type Adapter func(http.Handler) http.Handler

func InitContext() Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		})
	}
}

// func InitContextWith() Adapter {
// 	return func(h http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			r = r.WithContext() //...
// 			h.ServeHTTP(w, r)
// 		})
// 	}
// }
func InitLoggingResponseWriter() Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writer := newLoggingResponseWriter(w)

			h.ServeHTTP(writer, r)
		})
	}
}

func InitMetrics() Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		})
	}
}
