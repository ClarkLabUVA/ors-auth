package main

import (
	"net/http"
	"io/ioutil"
	"errors"
	"encoding/json"
)


func ResourceCreate(w http.ResponseWriter, r *http.Request) {

	// read and marshal body json into
	var res Resource
	var err error
	var requestBody []byte

	requestBody, err = ioutil.ReadAll(r.Body)

	// If Error for Unmarshaling JSON Body
	if !json.Valid(requestBody) {
		w.WriteHeader(400)
		w.Write([]byte(`{"error": "Invalid JSON Submitted"}`))
		return
	}

	err = json.Unmarshal(requestBody, &res)

	if err != nil {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"error": "Failed to Unmarshal Request JSON"}`))
		return
	}

	err = res.Create()

	if err == nil {
		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"created": {"@id": "` + res.Id + `"}}`))
		return
	}

	// Error for when the User with u.Id Already Exists
	if errors.Is(err, ErrDocumentExists) {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"error": "User Already Exists" ,"@id": "` + res.Id + `"}`))
		return
	}

	// Unknown Error catch all
	w.WriteHeader(500)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write([]byte(`{"error": "` + err.Error() + `"}`))
	return

}

// TODO: (LowPriority) Write Handler for basic Get Resource by ID
func ResourceGet(w http.ResponseWriter, r *http.Request) {}

// TODO: (MidPriority) Write Handler for deletion by ID
func ResourceDelete(w http.ResponseWriter, r *http.Request) {}

// TODO: (LowPriority) Write Handler listing and filtering resources
func ResourceList(w http.ResponseWriter, r *http.Request) {}
