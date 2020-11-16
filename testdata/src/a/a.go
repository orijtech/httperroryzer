// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
)

func badHandlerPlainNoReturn(r *http.Request, w http.ResponseWriter) {
	res, err := http.DefaultClient.Get("https://golang.org/non-existent")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // want "call to http.Error without a terminating statement below it"
	}
	defer res.Body.Close()
	_, _ = ioutil.ReadAll(res.Body)
}

func badHandlerWithLog(_ context.Context, r *http.Request, w http.ResponseWriter) {
	res, err := http.DefaultClient.Get("https://golang.org/non-existent")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // want "call to http.Error without a terminating statement below it"
		log.Printf("Ends here")
	}
	defer res.Body.Close()
	_, _ = ioutil.ReadAll(res.Body)
}

func goodMixedHandlerLogFatal(_ context.Context, r *http.Request, w http.ResponseWriter) {
	res, err := http.DefaultClient.Get("https://golang.org/non-existent")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatal("Ends here")
	}
	defer res.Body.Close()
	_, _ = ioutil.ReadAll(res.Body)
}

func goodMixedHandlerLogFatalf(_ context.Context, r *http.Request, w http.ResponseWriter) {
	res, err := http.DefaultClient.Get("https://golang.org/non-existent")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatalf("Ends here")
	}
	defer res.Body.Close()
	_, _ = ioutil.ReadAll(res.Body)
}

func goodMixedHandlerPanic(_ context.Context, r *http.Request, w http.ResponseWriter) {
	res, err := http.DefaultClient.Get("https://golang.org/non-existent")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		panic("Ends here")
	}
	defer res.Body.Close()
	_, _ = ioutil.ReadAll(res.Body)
}

func goodHandlerRuntimeGoexit(w http.ResponseWriter, r *http.Request) {
	res, err := http.DefaultClient.Get("https://golang.org/non-existent")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		runtime.Goexit()
	}
	defer res.Body.Close()
	_, _ = ioutil.ReadAll(res.Body)
}

func goodBare(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Unimplemented", http.StatusMethodNotAllowed)
}

func goodBareWithPrintf(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Unimplemented", http.StatusMethodNotAllowed)
	log.Println("Done here")
}

func goodWithTerminalStatement(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Unimplemented", http.StatusMethodNotAllowed)
	panic("Done here")
}

func goodBareUnimplementedScopePruning(w http.ResponseWriter, r *http.Request) {
	{
		http.Error(w, "Unimplemented", http.StatusMethodNotAllowed)
	}
	log.Println("Done here")
}

func unreachableHTTPError(w http.ResponseWriter, r *http.Request) {
	if true {
		http.Error(w, "Early end first", 400)
		return
		http.Error(w, "Unreachable", 200) // want "call to http.Error without a terminating statement below it"
	}

	println("Done")
}
