// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package http

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/swarm/api"
)

//metrics variables
var (
	htmlCounter = metrics.NewRegisteredCounter("api.http.errorpage.html.count", nil)
	jsonCounter = metrics.NewRegisteredCounter("api.http.errorpage.json.count", nil)
)

//parameters needed for formatting the correct HTML page
type ResponseParams struct {
	Msg       string
	Code      int
	Timestamp string
	template  *template.Template
	Details   template.HTML
}

//ShowMultipeChoices is used when a user requests a resource in a manifest which results
//in ambiguous results. It returns a HTML page with clickable links of each of the entry
//in the manifest which fits the request URI ambiguity.
//For example, if the user requests bzz:/<hash>/read and that manifest contains entries
//"readme.md" and "readinglist.txt", a HTML page is returned with this two links.
//This only applies if the manifest has no default entry
func ShowMultipleChoices(w http.ResponseWriter, r *http.Request, list api.ManifestList) {
	msg := ""
	if list.Entries == nil {
		RespondError(w, r, "Could not resolve", http.StatusInternalServerError)
		return
	}
	//make links relative
	//requestURI comes with the prefix of the ambiguous path, e.g. "read" for "readme.md" and "readinglist.txt"
	//to get clickable links, need to remove the ambiguous path, i.e. "read"
	idx := strings.LastIndex(r.RequestURI, "/")
	if idx == -1 {
		RespondError(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	//remove ambiguous part
	base := r.RequestURI[:idx+1]
	for _, e := range list.Entries {
		//create clickable link for each entry
		msg += "<a href='" + base + e.Path + "'>" + e.Path + "</a><br/>"
	}
	RespondTemplate(w, r, "multiple-choice", msg, http.StatusMultipleChoices)
}

func RespondTemplate(w http.ResponseWriter, r *http.Request, templateName, msg string, code int) {
	respond(w, r, &ResponseParams{
		Code:      code,
		Msg:       msg,
		Timestamp: time.Now().Format(time.RFC1123),
		template:  TemplatesMap[templateName],
	})
}

func RespondError(w http.ResponseWriter, r *http.Request, msg string, code int) {
	RespondTemplate(w, r, "error", msg, code)
}

//evaluate if client accepts html or json response
func respond(w http.ResponseWriter, r *http.Request, params *ResponseParams) {
	w.WriteHeader(params.Code)
	switch r.Header.Get("Accept") {
	case "application/json":
		respondJSON(w, params)

	case "text/plain":
		//curl

	default:
		respondHTML(w, params)
	}

}

//return a HTML page
func respondHTML(w http.ResponseWriter, params *ResponseParams) {
	htmlCounter.Inc(1)
	err := params.template.Execute(w, params)
	if err != nil {
		log.Error(err.Error())
	}
}

//return JSON
func respondJSON(w http.ResponseWriter, params *ResponseParams) {
	jsonCounter.Inc(1)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(params)
}
