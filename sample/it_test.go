package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthenticate(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{
			name:    "empty",
			wantErr: true,
		},
		{
			name:    "whitespace",
			secret:  " ",
			wantErr: true,
		},
		{
			name:    "mismatched secret",
			secret:  "not-the-secret",
			wantErr: true,
		},
		{
			name:   "proper secret",
			secret: "open-sesame",
		},
	}

	cst := httptest.NewServer(http.HandlerFunc(authenticate))
	defer cst.Close()

	client := cst.Client()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", cst.URL, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("x-secret", tt.secret)
			res, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			blob, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			switch tt.wantErr {
			case true:
				if g, w := res.StatusCode, http.StatusUnauthorized; g != w {
					t.Errorf("StatusCode mismatch:\nGot: %q\nWant statusCode: %d", res.Status, w)
				}
				if !bytes.Contains(blob, []byte("unauthorized")) {
					t.Errorf(`Missing "unauthorized", got %q`, blob)
				}

			default:
				if g, w := res.StatusCode, http.StatusOK; g != w {
					t.Errorf("StatusCode mismatch:\nGot:  %q\nWant: %q", g, w)
				}
				if !bytes.Contains(blob, []byte(`"secret_location":`)) {
					t.Errorf(`Body does not contain "secret_location", got %q`, blob)
				}
			}
		})
	}
}
