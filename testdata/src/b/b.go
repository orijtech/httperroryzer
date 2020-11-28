package a

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
)

func proxyHandleWithoutReturn(rw http.ResponseWriter, req *http.Request) {
	if err := do(); err != nil {
		if testing.Verbose() {
			fmt.Fprintf(os.Stderr, "go proxy: no archive: %v\n", err)
		}
		if errors.Is(err, os.ErrNotExist) {
			http.NotFound(rw, req) // want "call to http.Error without a terminating statement below it"
		} else {
			http.Error(rw, "cannot load archive", 500) // want "call to http.Error without a terminating statement below it"
		}
		if req.Header.Get("Content-Type") != "video/mp4" {
			if req.Header.Get("X-Auth-Token") != "open-sesame" {
				if req.Header.Get("Origin") != "/home" {
					http.Error(rw, "non-matching incarnation for the video", http.StatusBadRequest) // want "call to http.Error without a terminating.+"
				}
			}
		}
	}

	ext := req.Header.Get("x-ext")
	switch ext {
	case "zip":
		rw.Write([]byte("Zip here"))
	}
}

func proxyHandleWithReturn(rw http.ResponseWriter, req *http.Request) {
	if err := do(); err != nil {
		if testing.Verbose() {
			fmt.Fprintf(os.Stderr, "go proxy: no archive: %v\n", err)
		}
		if errors.Is(err, os.ErrNotExist) {
			http.NotFound(rw, req) // want "call to http.Error without a terminating statement below it"
		} else {
			http.Error(rw, "cannot load archive", 500) // want "call to http.Error without a terminating.+"
		}
		if req.Header.Get("Content-Type") != "video/mp4" {
			if req.Header.Get("X-Auth-Token") != "open-sesame" {
				if req.Header.Get("Origin") != "/home" {
					http.Error(rw, "non-matching incarnation for the video", http.StatusBadRequest)
				}
			}
		}
		return
	}

	ext := req.Header.Get("x-ext")
	switch ext {
	case "zip":
		rw.Write([]byte("Zip here"))
	}
}

func notFoundReplier(rw http.ResponseWriter, req *http.Request) {
	if req.Header.Get("failalways") != "" {
		http.NotFound(rw, req) // want "call to http.Error without a terminating.+"
	}
	rw.Write([]byte("FOO"))
}

func notFoundReplierWithReturn(rw http.ResponseWriter, req *http.Request) {
	if req.Header.Get("failalways") != "" {
		http.NotFound(rw, req)
		return
	}
	rw.Write([]byte("FOO"))
}

func foo(rw http.ResponseWriter, req *http.Request) {
	http.Error(rw, "msg", 404)
}

func notFoundReplierWithoutReturn(rw http.ResponseWriter, req *http.Request) {
	if req.Header.Get("failalways") != "" {
		http.NotFound(rw, req)
	}
}

func notFoundReplierWithBranch(rw http.ResponseWriter, req *http.Request) {
	if req.Header.Get("failalways") != "" {
		rw.WriteHeader(200)
		http.NotFound(rw, req)
		rw.Write([]byte("fizz buzz"))
	} else {
		http.Error(rw, "ditto", http.StatusBadRequest)
	}
}

func do() error {
	return errors.New("foo")
}
