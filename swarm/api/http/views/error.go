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

/*
Show nicely (but simple) formatted HTML error pages (or respond with JSON
if the appropriate `Accept` header is set)) for the http package.
*/
package views

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/swarm/api/http/messages"
)

//templateMap holds a mapping of an HTTP error code to a template
var templateMap map[int]*template.Template
var caseErrors []CaseError

//a custom error case struct that would be used to store validators and
//additional error info to display with client responses.
type CaseError struct {
	Validator func(*messages.Request) bool
	Msg       func(*messages.Request) string
}

//we init the error handling right on boot time, so lookup and http response is fast
func init() {
	initErrHandling()
}

func initErrHandling() {
	//pages are saved as strings - get these strings
	genErrPage := GetGenericErrorPage()
	notFoundPage := GetNotFoundErrorPage()
	multipleChoicesPage := GetMultipleChoicesErrorPage()
	//map the codes to the available pages
	tnames := map[int]string{
		0: genErrPage, //default
		http.StatusBadRequest:          genErrPage,
		http.StatusNotFound:            notFoundPage,
		http.StatusMultipleChoices:     multipleChoicesPage,
		http.StatusInternalServerError: genErrPage,
	}
	templateMap = make(map[int]*template.Template)
	for code, tname := range tnames {
		//assign formatted HTML to the code
		templateMap[code] = template.Must(template.New(fmt.Sprintf("%d", code)).Parse(tname))
	}

	caseErrors = []CaseError{
		{
			Validator: func(r *messages.Request) bool {
				return r.Uri != nil && r.Uri.Addr != "" && strings.HasPrefix(r.Uri.Addr, "0x")
			},
			Msg: func(r *messages.Request) string {
				uriCopy := r.Uri
				uriCopy.Addr = strings.TrimPrefix(uriCopy.Addr, "0x")
				return fmt.Sprintf(`The requested hash seems to be prefixed with '0x'. You will be redirected to the correct URL within 5 seconds.<br/>
			Please click <a href='%[1]s'>here</a> if your browser does not redirect you.<script>setTimeout("location.href='%[1]s';",5000);</script>`, "/"+uriCopy.String())
			},
		}}
}

//ValidateCaseErrors is a method that process the request object through certain validators
//that assert if certain conditions are met for further information to log as an error
func ValidateCaseErrors(r *messages.Request) string {
	for _, err := range caseErrors {
		if err.Validator(r) {
			return err.Msg(r)
		}
	}

	return ""
}

//get the HTML template for a given code
func GetTemplate(code int) *template.Template {
	if val, tmpl := templateMap[code]; tmpl {
		return val
	}
	return templateMap[0]
}
