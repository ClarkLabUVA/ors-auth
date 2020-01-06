package main

import (
	"net/http"
	"io/ioutil"
	"errors"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
)

// POST /user/
func UserCreate(w http.ResponseWriter, r *http.Request) {

	// read and marshal body json into
	var u User
	var err error
	var requestBody []byte

	requestBody, err = ioutil.ReadAll(r.Body)

	// If Error for Unmarshaling JSON Body
	if !json.Valid(requestBody) {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"error": "Invalid JSON Submitted"}`))
		return
	}

	err = json.Unmarshal(requestBody, &u)

	if err != nil {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"message": "Failed to Unmarshal Request JSON", "error": "` + err.Error() + `"}`))
		return
	}

	err = u.Create()

	if err == nil {
		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"created": {"@id": "` + u.Id + `"}}`))
		return
	}

	// Error for when the User with u.Id Already Exists
	if errors.Is(err, ErrDocumentExists) {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"error": "User Already Exists" ,"@id": "` + u.Id + `"}`))
		return
	}

	// Unknown Error catch all
	w.WriteHeader(500)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write([]byte(`{"error": "` + err.Error() + `"}`))
	return

}

// GET /user/
func UserList(w http.ResponseWriter, r *http.Request) {

	var err error
	var userList []User

	userList, err = listUsers()

	if err != nil {
		return
	}

	var responseBody []byte
	responseBody, err = json.Marshal(userList)

	if err != nil {
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write(responseBody)

	return
}

// GET /user/:userID
func UserGet(w http.ResponseWriter, r *http.Request) {
	var u User
	var err error

	// get the user id from the route
	params := httprouter.ParamsFromContext(r.Context())

	u.Id = params.ByName("userID")
	err = u.Get()

	if err != nil {
		return
	}

	responseBytes, err := json.Marshal(u)

	if err != nil {
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write(responseBytes)
	return

}

// Delete /user/:userID
func UserDelete(w http.ResponseWriter, r *http.Request) {

	var deletedUser User
	var err error

	// get the user id from the route
	params := httprouter.ParamsFromContext(r.Context())

	deletedUser.Id = params.ByName("userID")
	err = deletedUser.Delete()

	if err != nil {
		return
	}

	responseBytes, err := json.Marshal(deletedUser)
	if err != nil {
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write(responseBytes)
	return

}
