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

func (s *Server) HandleBzzGet(w http.ResponseWriter, r *Request) {
	log.Debug("handleGetBzz")
	if r.Header.Get("Accept") == "application/x-tar" {
		reader, err := s.api.GetDirectoryTar(r.Context(), r.uri)
		if err != nil {
			Respond(w, r, fmt.Sprintf("Had an error building the tarball: %v", err), http.StatusInternalServerError)
		}
		defer reader.Close()

		w.Header().Set("Content-Type", "application/x-tar")
		w.WriteHeader(http.StatusOK)
		io.Copy(w, reader)
		return
	}
	s.HandleGetFile(w, r)
}

func (s *Server) HandleBzzPost(w http.ResponseWriter, r *Request) {
	log.Debug("handlePostFiles")
	s.HandlePostFiles(w, r)
}

func (s *Server) HandleBzzDelete(w http.ResponseWriter, r *Request) {
	log.Debug("handleBzzDelete")
	s.HandleDelete(w, r)
}

func (s *Server) HandleBzzRawGet(w http.ResponseWriter, r *Request) {
	log.Debug("handleGetRaw")
	s.HandleGet(w, r)
}

func (s *Server) HandleBzzRawPost(w http.ResponseWriter, r *Request) {
	log.Debug("handlePostRaw")
	s.HandlePostRaw(w, r)
}

func (s *Server) HandleBzzImmutableGet(w http.ResponseWriter, r *Request) {
	log.Debug("handleGetHash")
	s.HandleGetList(w, r)
}

func (s *Server) HandleBzzHashGet(w http.ResponseWriter, r *Request) {
	log.Debug("handleGetHash")
	s.HandleGet(w, r)
}

func (s *Server) HandleBzzListGet(w http.ResponseWriter, r *Request) {
	log.Debug("handleGetHash")
	s.HandleGetList(w, r)
}

func (s *Server) HandleBzzResourceGet(w http.ResponseWriter, r *Request) {
	log.Debug("handleGetResource")
	s.HandleGetResource(w, r)
}

func (s *Server) HandleBzzResourcePost(w http.ResponseWriter, r *Request) {
	log.Debug("handlePostResource")
	s.HandlePostResource(w, r)
}

func (s *Server) HandleRootPaths(w http.ResponseWriter, r *Request) {
	switch r.Method {
	case http.MethodGet:
		if r.RequestURI == "/" {
			if strings.Contains(r.Header.Get("Accept"), "text/html") {
				err := landingPageTemplate.Execute(w, nil)
				if err != nil {
					log.Error(fmt.Sprintf("error rendering landing page: %s", err))
				}
				return
			}
			if strings.Contains(r.Header.Get("Accept"), "application/json") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode("Welcome to Swarm!")
				return
			}
		}

		if r.URL.Path == "/robots.txt" {
			w.Header().Set("Last-Modified", time.Now().Format(http.TimeFormat))
			fmt.Fprintf(w, "User-agent: *\nDisallow: /")
			return
		}
		Respond(w, r, "Bad Request", http.StatusBadRequest)
	default:
		Respond(w, r, "Not Found", http.StatusNotFound)
	}
}
