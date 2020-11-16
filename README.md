# httperroryzer
Static analyzer to catch invalid uses of http.Error without a return statement which can cause expected bugs

## Example
An insidious bug could be the following in an HTTP handler

```go
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
```

In here, the caller forgot to invoke return after every error and in the end exposed what was supposed to be privileged
information only available to those with the right credentials, for example when ran as per https://play.golang.org/p/I7Swd6GBfYe,
it'll print out with malformed/malicious input such as `{` or `{"secret":""}`
```shell 
http: superfluous response.WriteHeader call from main.authenticate (secrets.go:45)
could not parse out the response
unauthorized
{"secret_location":"23.4162° N, 25.6628° E"}
```

### Remedy
Let's what would happen if we ran httperroryzer on it
```shell
go get github.com/orijtech/httperroryzer/cmd/httperroryzer && httperroryzer secrets.go

/go/src/github.com/orijtech/httperroryzer/sample/secrets.go:42:3: call to http.Error without a terminating statement below it
/go/src/github.com/orijtech/httperroryzer/sample/secrets.go:45:3: call to http.Error without a terminating statement below it
```

and when corrected
```shell
func authenticate(rw http.ResponseWriter, req *http.Request) {
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	creds := new(credentials)
	if err := dec.Decode(creds); err != nil {
		http.Error(rw, "could not parse out the response", http.StatusBadRequest)
		return
	}
	if creds.Secret != "open-sesame" {
		http.Error(rw, "unauthorized", http.StatusUnauthorized)
		return
	}
	rw.Write([]byte(`{"secret_location":"23.4162° N, 25.6628° E"}`))
}
```
