package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/swarm/log"
)

func (s *Server) HandleBzzGet() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("handleGetBzz")
		if r.Header.Get("Accept") == "application/x-tar" {
			uri := GetURI(r.Context())
			reader, err := s.api.GetDirectoryTar(r.Context(), uri)
			if err != nil {
				RespondError(w, r, fmt.Sprintf("Had an error building the tarball: %v", err), http.StatusInternalServerError)
			}
			defer reader.Close()

			w.Header().Set("Content-Type", "application/x-tar")
			w.WriteHeader(http.StatusOK)
			io.Copy(w, reader)
			return
		}

		s.HandleGetFile(w, r)
	})
}

func (s *Server) HandleBzzPost() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("handlePostFiles")
		s.HandlePostFiles(w, r)
	})
}

func (s *Server) HandleBzzDelete() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("handleBzzDelete")
		s.HandleDelete(w, r)
	})
}

func (s *Server) HandleBzzRawGet() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("handleGetRaw")
		s.HandleGet(w, r)
	})
}

func (s *Server) HandleBzzRawPost() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("handlePostRaw")
		s.HandlePostRaw(w, r)
	})
}

func (s *Server) HandleBzzImmutableGet() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("handleGetHash")
		s.HandleGetList(w, r)
	})
}

func (s *Server) HandleBzzHashGet() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("handleGetHash")
		s.HandleGet(w, r)
	})
}

func (s *Server) HandleBzzListGet() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("handleGetHash")
		s.HandleGetList(w, r)
	})
}

func (s *Server) HandleBzzResourceGet() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("handleGetResource")
		s.HandleGetResource(w, r)
	})
}

func (s *Server) HandleBzzResourcePost() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("handlePostResource")
		s.HandlePostResource(w, r)
	})
}

func (s *Server) HandleRootPaths() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/":
			if strings.Contains(r.Header.Get("Accept"), "text/html") {
				RespondTemplate(w, r, "landing-page", "", 200)
			}
			if strings.Contains(r.Header.Get("Accept"), "application/json") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode("Welcome to Swarm!")
			}
		case "/robots.txt":
			w.Header().Set("Last-Modified", time.Now().Format(http.TimeFormat))
			fmt.Fprintf(w, "User-agent: *\nDisallow: /")
		case "/favicon.ico":
			w.WriteHeader(http.StatusOK)
			w.Write(faviconBytes)
		default:
			RespondError(w, r, "Not Found", http.StatusNotFound)
		}
	})
}
