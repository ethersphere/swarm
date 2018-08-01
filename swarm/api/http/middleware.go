package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/pborman/uuid"
)

func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range adapters {
		h = adapter(h)
	}
	return HeaderControl(h)
}

type Adapter func(http.Handler) http.Handler

func SetRequestID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(SetRUID(r.Context(), uuid.New()[:8]))
		metrics.GetOrRegisterCounter(fmt.Sprintf("http.request.%s", r.Method), nil).Inc(1)
		log.Info("created ruid for request", "ruid", GetRUID(r.Context()), "method", r.Method, "url", r.RequestURI)

		h.ServeHTTP(w, r)
	})
}

func ParseURI(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri, err := api.Parse(strings.TrimLeft(r.URL.Path, "/"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			RespondError(w, r, fmt.Sprintf("invalid URI %q", r.URL.Path), http.StatusBadRequest)
			return
		}
		if uri.Addr != "" && strings.HasPrefix(uri.Addr, "0x") {
			uri.Addr = strings.TrimPrefix(uri.Addr, "0x")

			msg := fmt.Sprintf(`The requested hash seems to be prefixed with '0x'. You will be redirected to the correct URL within 5 seconds.<br/>
			Please click <a href='%[1]s'>here</a> if your browser does not redirect you within 5 seconds.<script>setTimeout("location.href='%[1]s';",5000);</script>`, "/"+uri.String())
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(msg))
			return
		}

		ctx := r.Context()
		r = r.WithContext(SetURI(ctx, uri))

		log.Debug("parsed request path", "ruid", GetRUID(r.Context()), "method", r.Method, "uri.Addr", uri.Addr, "uri.Path", uri.Path, "uri.Scheme", uri.Scheme)

		h.ServeHTTP(w, r)
	})
}

func InitLoggingResponseWriter(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := newLoggingResponseWriter(w)

		h.ServeHTTP(writer, r)
	})
}

func InitMetrics(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

func HeaderControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if w.Header().Get("Status") == "" {
				return
			}
			code, err := strconv.Atoi(w.Header().Get("Status"))
			if err != nil {
				log.Error("had an error parsing the string to int")
				panic(err)
			}
			if code >= 400 {
				w.Header().Del("Cache-Control") //avoid sending cache headers for errors!
				w.Header().Del("ETag")
			}
		}()
		h.ServeHTTP(w, r)
	})
}
