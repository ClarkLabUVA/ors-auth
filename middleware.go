package main

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
)

func BasicAuth(h httprouter.Handle, requiredUser, requiredPassword string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Get the Basic Authentication credentials
		user, password, hasAuth := r.BasicAuth()

		if hasAuth && user == requiredUser && password == requiredPassword {
			// Delegate request to the given handle
			h(w, r, ps)
		} else {
			// Request Basic Authentication otherwise
			w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	}
}

func ValidJSON(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	var err error
	var requestBody []byte

	requestBody, err = ioutil.ReadAll(r.Body)

	if err != nil {
		rw.WriteHeader(400)
		rw.Header().Set("Content-Type", "application/ld+json")
		rw.Write([]byte(`{"error": "Unable to Read Request Body"}`))
		return
	}

	// If Error for Unmarshaling JSON Body
	if !json.Valid(requestBody) {
		rw.WriteHeader(400)
		rw.Header().Set("Content-Type", "application/ld+json")
		rw.Write([]byte(`{"error": "Invalid JSON Submitted"}`))
		return
	}

	next(rw, r)
}

func DefaultMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	rw.Header().Set("Content-Type", "application/ld+json")
	next(rw, r)
}
