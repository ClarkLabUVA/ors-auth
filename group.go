package main

import (
	"net/http"
	"errors"
	"io/ioutil"
	"encoding/json"
)


func GroupCreate(w http.ResponseWriter, r *http.Request) {

	// read and marshal body json into
	var g Group
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

	err = g.Create()

	if err == nil {
		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"created": {"@id": "` + g.Id + `"}}`))
		return
	}

	// Error for when the User with u.Id Already Exists
	if errors.Is(err, ErrDocumentExists) {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"error": "User Already Exists" ,"@id": "` + g.Id + `"}`))
		return
	}

	// Unknown Error catch all
	w.WriteHeader(500)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write([]byte(`{"error": "` + err.Error() + `"}`))
	return

}

// TODO: (LowPriority) Write Endpoint GroupGet
func GroupGet(w http.ResponseWriter, r *http.Request) {}

// TODO: (LowPriority) Write Endpoint GroupUpdate
func GroupUpdate(w http.ResponseWriter, r *http.Request) {}

// TODO: (LowPriority) Write Endpoint GroupDelete
func GroupDelete(w http.ResponseWriter, r *http.Request) {}

// TODO: (LowPriority) Write Endpoint GroupList
func GroupList(w http.ResponseWriter, r *http.Request) {}
