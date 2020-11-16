package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidity(t *testing.T) {
}

func main() {
	cst := httptest.NewServer(http.HandlerFunc(authenticate))
	defer cst.Close()

	client := cst.Client()
	req, err := http.NewRequest("POST", cst.URL, strings.NewReader(`{`))
	if err != nil {
		panic(err)
	}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	location, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	println(string(location))
}

type credentials struct {
	Secret string `json:"secret"`
}

func authenticate(rw http.ResponseWriter, req *http.Request) {
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	creds := new(credentials)
	if err := dec.Decode(creds); err != nil {
		http.Error(rw, "could not parse out the response", http.StatusBadRequest)
	}
	if creds.Secret != "open-sesame" {
		http.Error(rw, "unauthorized", http.StatusUnauthorized)
	}
	rw.Write([]byte(`{"secret_location":"23.4162° N, 25.6628° E"}`))
}
