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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/net/html"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/testutil"
)

func TestError(t *testing.T) {
	srv := testutil.NewTestSwarmServer(t, serverFunc, nil)
	defer srv.Close()

	var resp *http.Response
	var respbody []byte

	url := srv.URL + "/this_should_fail_as_no_bzz_protocol_present"
	resp, err := http.Get(url)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	respbody, err = ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 404 && !strings.Contains(string(respbody), "Invalid URI &#34;/this_should_fail_as_no_bzz_protocol_present&#34;: unknown scheme") {
		t.Fatalf("Response body does not match, expected: %v, to contain: %v; received code %d, expected code: %d", string(respbody), "Invalid bzz URI: unknown scheme", 400, resp.StatusCode)
	}

	_, err = html.Parse(strings.NewReader(string(respbody)))
	if err != nil {
		t.Fatalf("HTML validation failed for error page returned!")
	}
}

func Test404Page(t *testing.T) {
	srv := testutil.NewTestSwarmServer(t, serverFunc, nil)
	defer srv.Close()

	var resp *http.Response
	var respbody []byte

	url := srv.URL + "/bzz:/1234567890123456789012345678901234567890123456789012345678901234"
	resp, err := http.Get(url)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	respbody, err = ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 404 || !strings.Contains(string(respbody), "404") {
		t.Fatalf("Invalid Status Code received, expected 404, got %d", resp.StatusCode)
	}

	_, err = html.Parse(strings.NewReader(string(respbody)))
	if err != nil {
		t.Fatalf("HTML validation failed for error page returned!")
	}
}

func Test500Page(t *testing.T) {
	srv := testutil.NewTestSwarmServer(t, serverFunc, nil)
	defer srv.Close()

	var resp *http.Response
	var respbody []byte

	url := srv.URL + "/bzz:/thisShouldFailWith500Code"
	resp, err := http.Get(url)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	respbody, err = ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 404 {
		t.Fatalf("Invalid Status Code received, expected 404, got %d", resp.StatusCode)
	}

	_, err = html.Parse(strings.NewReader(string(respbody)))
	if err != nil {
		t.Fatalf("HTML validation failed for error page returned!")
	}
}
func Test500PageWith0xHashPrefix(t *testing.T) {
	srv := testutil.NewTestSwarmServer(t, serverFunc, nil)
	defer srv.Close()

	var resp *http.Response
	var respbody []byte

	url := srv.URL + "/bzz:/0xthisShouldFailWith500CodeAndAHelpfulMessage"
	resp, err := http.Get(url)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	respbody, err = ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 404 {
		t.Fatalf("Invalid Status Code received, expected 404, got %d", resp.StatusCode)
	}

	if !strings.Contains(string(respbody), "The requested hash seems to be prefixed with") {
		t.Fatalf("Did not receive the expected error message")
	}

	_, err = html.Parse(strings.NewReader(string(respbody)))
	if err != nil {
		t.Fatalf("HTML validation failed for error page returned!")
	}
}

func TestJsonResponse(t *testing.T) {
	srv := testutil.NewTestSwarmServer(t, serverFunc, nil)
	defer srv.Close()

	var resp *http.Response
	var respbody []byte

	url := srv.URL + "/bzz:/thisShouldFailWith500Code/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	defer resp.Body.Close()
	respbody, err = ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 404 {
		t.Fatalf("Invalid Status Code received, expected 404, got %d", resp.StatusCode)
	}

	if !isJSON(string(respbody)) {
		t.Fatalf("Expected response to be JSON, received invalid JSON: %s", string(respbody))
	}
}

func TestGetFallbackToList(t *testing.T) {
	srv := testutil.NewTestSwarmServer(t, serverFunc, nil)
	defer srv.Close()
	data := "arbitraryString"
	url := fmt.Sprintf("%s/bzz:/", srv.URL)

	buf := new(bytes.Buffer)
	form := multipart.NewWriter(buf)
	form.WriteField("name", "John Doe")
	file1, _ := form.CreateFormFile("cv", "cv.txt")
	file1.Write([]byte(data))
	file2, _ := form.CreateFormFile("profile_picture", "profile.jpg")
	file2.Write([]byte(data))
	form.Close()

	headers := map[string]string{
		"Content-Type":   form.FormDataContentType(),
		"Content-Length": strconv.Itoa(buf.Len()),
	}
	res, body := httpDo("POST", url, buf, headers, false, t)

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected POST multipart/form-data to return 200, but it returned %d", res.StatusCode)
	}
	if len(body) != 64 {
		t.Fatalf("expected POST multipart/form-data to return a 64 char manifest but the answer was %d chars long", len(body))
	}
	log.Info(fmt.Sprintf("uploading directory with 'swarm up'"))
	hash := body
	log.Info("dir uploaded", "hash", hash)
	headers = map[string]string{"Accept": "*/*"}
	res, body = httpDo("GET", srv.URL+"/bzz:/"+hash, nil, headers, false, t)
	if res.StatusCode != 301 {
		//		t.Fatalf("expected HTTP status 301, got %s", res.Status)
	}
	//todo: check the location header
	if body != data {
		t.Fatalf("expected HTTP body %q, got %q", data, body)
	}
}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
