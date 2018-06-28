package controllers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/api/http/messages"
	"github.com/ethereum/go-ethereum/swarm/api/http/views"
	l "github.com/ethereum/go-ethereum/swarm/log"
)

//metrics variables
var (
	htmlCounter = metrics.NewRegisteredCounter("api.http.errorpage.html.count", nil)
	jsonCounter = metrics.NewRegisteredCounter("api.http.errorpage.json.count", nil)
)

type Controller struct {
	ControllerHandler
}

type ControllerHandler interface {
	Get(w http.ResponseWriter, r *messages.Request)
	// Post(w http.ResponseWriter, r *messages.Request)
	// Put(w http.ResponseWriter, r *messages.Request)
	// Delete(w http.ResponseWriter, r *messages.Request)
	// Patch(w http.ResponseWriter, r *messages.Request)
	Respond(w http.ResponseWriter, req *messages.Request, msg string, code int)
}

//Respond is used to show an HTML page to a client.
//If there is an `Accept` header of `application/json`, JSON will be returned instead
//The function just takes a string message which will be displayed in the error page.
//The code is used to evaluate which template will be displayed
//(and return the correct HTTP status code)
func (controller *Controller) Respond(w http.ResponseWriter, req *messages.Request, msg string, code int) {
	additionalMessage := views.ValidateCaseErrors(req)
	switch code {
	case http.StatusInternalServerError:
		log.Output(msg, log.LvlError, l.CallDepth, "ruid", req.Ruid, "code", code)
	default:
		log.Output(msg, log.LvlDebug, l.CallDepth, "ruid", req.Ruid, "code", code)
	}

	if code >= 400 {
		w.Header().Del("Cache-Control") //avoid sending cache headers for errors!
		w.Header().Del("ETag")
	}

	respond(w, &req.Request, &messages.ResponseParams{
		Code:      code,
		Msg:       msg,
		Details:   template.HTML(additionalMessage),
		Timestamp: time.Now().Format(time.RFC1123),
		Template:  views.GetTemplate(code),
	})
}

//evaluate if client accepts html or json response
func respond(w http.ResponseWriter, r *http.Request, params *messages.ResponseParams) {
	w.WriteHeader(params.Code)
	if r.Header.Get("Accept") == "application/json" {
		RespondJSON(w, params)
	} else {
		RespondHTML(w, params)
	}
}

//return a HTML page
func RespondHTML(w http.ResponseWriter, params *messages.ResponseParams) {
	htmlCounter.Inc(1)
	err := params.Template.Execute(w, params)
	if err != nil {
		log.Error(err.Error())
	}
}

//return JSON
func RespondJSON(w http.ResponseWriter, params *messages.ResponseParams) {
	jsonCounter.Inc(1)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(params)
}

//ShowMultipeChoices is used when a user requests a resource in a manifest which results
//in ambiguous results. It returns a HTML page with clickable links of each of the entry
//in the manifest which fits the request URI ambiguity.
//For example, if the user requests bzz:/<hash>/read and that manifest contains entries
//"readme.md" and "readinglist.txt", a HTML page is returned with this two links.
//This only applies if the manifest has no default entry
func (controller *Controller) ShowMultipleChoices(w http.ResponseWriter, req *messages.Request, list api.ManifestList) {
	msg := ""
	if list.Entries == nil {
		controller.Respond(w, req, "Could not resolve", http.StatusInternalServerError)
		return
	}
	//make links relative
	//requestURI comes with the prefix of the ambiguous path, e.g. "read" for "readme.md" and "readinglist.txt"
	//to get clickable links, need to remove the ambiguous path, i.e. "read"
	idx := strings.LastIndex(req.RequestURI, "/")
	if idx == -1 {
		controller.Respond(w, req, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	//remove ambiguous part
	base := req.RequestURI[:idx+1]
	for _, e := range list.Entries {
		//create clickable link for each entry
		msg += "<a href='" + base + e.Path + "'>" + e.Path + "</a><br/>"
	}
	controller.Respond(w, req, msg, http.StatusMultipleChoices)
}
