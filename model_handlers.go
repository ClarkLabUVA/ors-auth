package main

import (
	"net/http"
	"errors"
)


// Record interface for all Model Types
// Reduces Boilerplate Error Handling in Basic Paths
// still returning an error for more complex cases for each handler
type Record interface {
	ID() string
	Get() error
	Create() error
	Delete() error
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
}

func HandleGet(rec Record, w http.ResponseWriter) error {
	err := rec.Get()

	// Boilerplate Error Handling
	// For Document Not Found return a 404 message
	if err != nil {

		if errors.Is(err, ErrNoDocument) {
			w.WriteHeader(404)
			w.Write([]byte(`{"@id": "`+ rec.ID() +`", "error": "No Record Found"}`))
			return nil
		}

		// return for more complex error handling back to the handler
		return err
	}

	HandleMarshal(rec, w)
	return nil
}

func HandleDelete(rec Record, w http.ResponseWriter) error {
	err := rec.Delete()

	if err != nil {
		// Handle Errors
		if errors.Is(err, ErrNoDocument) {
			w.WriteHeader(404)
			w.Write([]byte(`{"@id": "`+ rec.ID() +`", "error": "No Record Found"}`))
			return nil
		}

		return err
	}

	return nil
}

func HandleCreate(rec Record, w http.ResponseWriter) error {
	err := rec.Create()

	if err != nil {
		// Handle Errors
		if errors.Is(err, ErrDocumentExists) {
			w.WriteHeader(400)
			w.Write([]byte(`{"@id": "`+ rec.ID() +`", "error": "Document Already Exists"}`))
			return nil
		}

		return err
	}


	HandleMarshal(rec, w)
	return nil
}

func HandleMarshal(rec Record, w http.ResponseWriter) {

	responseBody, err := rec.MarshalJSON()

	// handle json marshaling error
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(`{"message": "Failed to Marshal JSON", "error": "` + err.Error() +`"}`))
	}

	// marshal response and return
	w.WriteHeader(200)
	w.Write(responseBody)

}
