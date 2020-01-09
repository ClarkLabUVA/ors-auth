package main

import (
	"testing"
	"errors"
	mongo "go.mongodb.org/mongo-driver/mongo"
)

func TestErrorDocuments(t *testing.T) {

	err := errors.New("Not Real")



	if ErrorDocumentExists(err) {
		t.Errorf("Error is not Document Exists")
	}

	mongoErr := mongo.WriteErrors{ mongo.WriteError{Code: 11000}}

	if !ErrorDocumentExists(mongoErr) {
		t.Errorf("ErrorDocumentExists fails to detect correct error ")
	}

}
