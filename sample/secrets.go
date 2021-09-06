package main

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
)

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
	blob, _ := httputil.DumpResponse(res, true)
	res.Body.Close()
	println(string(blob))
}

func authenticate(rw http.ResponseWriter, req *http.Request) {
	secret := req.Header.Get("x-secret")
	if secret != "open-sesame" {
		http.Error(rw, "unauthorized, please set header", http.StatusUnauthorized)
	}
	_, _ = rw.Write([]byte(`{"secret_location":"23.4162° N, 25.6628° E"}`))
}
