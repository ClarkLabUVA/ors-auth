package main

import (
	"net/http"
	"errors"
	"io/ioutil"
	"encoding/json"
)


func PolicyCreate(w http.ResponseWriter, r *http.Request) {

	// read and marshal body json into
	var p Policy
	var err error
	var requestBody []byte

	requestBody, err = ioutil.ReadAll(r.Body)

	// If Error for Unmarshaling JSON Body
	if !json.Valid(requestBody) {
		w.WriteHeader(400)
		w.Write([]byte(`{"error": "Invalid JSON Submitted"}`))
		return
	}

	err = json.Unmarshal(requestBody, &r)

	if err != nil {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"error": "Failed to Unmarshal Request JSON"}`))
		return
	}

	err = p.Create()

	if err == nil {
		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"created": {"@id": "` + p.Id + `"}}`))
		return
	}

	// Error for when the User with u.Id Already Exists
	if errors.Is(err, ErrDocumentExists) {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"error": "User Already Exists" ,"@id": "` + p.Id + `"}`))
		return
	}

	// Unknown Error catch all
	w.WriteHeader(500)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write([]byte(`{"error": "` + err.Error() + `"}`))
	return

}

// TODO: (LowPriority) Write Handler GetPolicy
func PolicyGet(w http.ResponseWriter, r *http.Request) {}

// TODO: (LowPriority) Write Handler PolicyUpdate
func PolicyUpdate(w http.ResponseWriter, r *http.Request) {}

// TODO: (LowPriority) Write Handler PolicyDelete
func PolicyDelete(w http.ResponseWriter, r *http.Request) {}

// TODO: (LowPriority) Write Handler PolicyList; filter by principal or resource as query params
func PolicyList(w http.ResponseWriter, r *http.Request) {}
