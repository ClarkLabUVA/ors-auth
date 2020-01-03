package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {

	router := httprouter.New()

	var globusClientID = flag.String("clientID", "", "Client ID from Globus Auth")
	var globusClientSecret = flag.String("clientSecret", "", "Client Secret from Globus Auth")

	var certPath = flag.String("cert", "", "Path to TLS Certificate")
	var keyPath = flag.String("key", "", "Path to TLS Key")

	var redirectURL = flag.String("redirect", "https://localhost:8080/oauth/token", "Redirect URL for Globus to Return to after access Code is granted")
	var scopes = "urn:globus:auth:scope:auth.globus.org:view_identity_set+urn:globus:auth:scope:auth.globus.org:view_identities+openid+email+profile"

	flag.Parse()

	if *globusClientID == "" || *globusClientSecret == "" {
		log.Fatalln("GlobusCredentials are  required")
	}

	if *certPath == "" || *keyPath == "" {
		log.Fatalln("TLS certificate and Key are Required")
	}

	globusClient := GlobusAuthClient{
		ClientID:     *globusClientID,
		ClientSecret: *globusClientSecret,
		RedirectURL:  *redirectURL,
		Scopes:       scopes,
	}

	router.Handler("GET", "/oauth/login", http.HandlerFunc(globusClient.GrantHandler))
	router.Handler("GET", "/oauth/token", http.HandlerFunc(globusClient.CodeHandler))
	router.Handler("POST", "/oauth/revoke", http.HandlerFunc(globusClient.RevokeHandler))

	//router.Handler("POST", "/oauth/register", http.HandlerFunc(globusClient.RegisterHandler))
	//router.Handler("POST", "/oauth/refresh", http.HandlerFunc(globusClient.RefreshHandler))

	router.Handler("POST", "/user", http.HandlerFunc(UserCreate))
	router.Handler("GET", "/user", http.HandlerFunc(UserList))
	router.Handler("GET", "/user/:userID", http.HandlerFunc(UserGet))
	router.Handler("DELETE", "/user/:userID", http.HandlerFunc(UserDelete))

	router.Handler("POST", "/challenge", http.HandlerFunc(ChallengeEvaluate))
	router.Handler("GET", "/challenge", http.HandlerFunc(ChallengeList))

	router.Handler("POST", "/resource", http.HandlerFunc(ResourceCreate))
	router.Handler("GET", "/resource", http.HandlerFunc(ResourceList))
	router.Handler("GET", "/resource/:ID", http.HandlerFunc(ResourceGet))
	router.Handler("DELETE", "/resource/:ID", http.HandlerFunc(ResourceDelete))

	router.Handler("POST", "/policy", http.HandlerFunc(PolicyCreate))
	router.Handler("GET", "/policy", http.HandlerFunc(PolicyList))
	router.Handler("GET", "/policy/:ID", http.HandlerFunc(PolicyGet))
	router.Handler("PUT", "/policy/:ID", http.HandlerFunc(PolicyUpdate))
	router.Handler("DELETE", "/policy/:ID", http.HandlerFunc(PolicyDelete))

	router.Handler("POST", "/group", http.HandlerFunc(GroupCreate))
	router.Handler("GET", "/group", http.HandlerFunc(GroupList))
	router.Handler("GET", "/group/:ID", http.HandlerFunc(GroupGet))
	router.Handler("PUT", "/group/:ID", http.HandlerFunc(GroupUpdate))
	router.Handler("DELETE", "/group/:ID", http.HandlerFunc(GroupDelete))

	log.Fatal(http.ListenAndServeTLS("0.0.0.0:8080", *certPath, *keyPath, router))

}

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

func ChallengeEvaluate(w http.ResponseWriter, r *http.Request) {

	var c Challenge
	var err error
	var requestBody []byte

	requestBody, err = ioutil.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error": "%w"}`, err)
		return
	}

	if err = json.Unmarshal(requestBody, &c); err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error": "%w"}`, err)
		return
	}

	err = c.Evaluate()
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error": "%w"}`, err)
		return
	}

	if c.Granted {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(403)
	}

	fmt.Fprintf(w, `{"@id": "%s","granted": "%t"}`, c.Id, c.Granted)
}

func ChallengeList(w http.ResponseWriter, r *http.Request) {

	var challengeList []Challenge
	var responseBody []byte
	var err error

	w.Header().Set("Content-Type", "application/ld+json")

	challengeList, err = listChallenges()

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, `{"error": "%w"}`, err)
		return
	}

	responseBody, err = json.Marshal(challengeList)

	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error": "%w"}`, err)
		return
	}

	w.WriteHeader(200)
	w.Write(responseBody)

}

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

func ResourceGet(w http.ResponseWriter, r *http.Request) {}

func ResourceDelete(w http.ResponseWriter, r *http.Request) {}

func ResourceList(w http.ResponseWriter, r *http.Request) {}

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

func PolicyGet(w http.ResponseWriter, r *http.Request) {}

func PolicyUpdate(w http.ResponseWriter, r *http.Request) {}

func PolicyDelete(w http.ResponseWriter, r *http.Request) {}

func PolicyList(w http.ResponseWriter, r *http.Request) {}

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

func GroupGet(w http.ResponseWriter, r *http.Request) {}

func GroupUpdate(w http.ResponseWriter, r *http.Request) {}

func GroupDelete(w http.ResponseWriter, r *http.Request) {}

func GroupList(w http.ResponseWriter, r *http.Request) {}

// Write a Response, & StatusCode for Any Application Error
func handleErrors(err error, w http.ResponseWriter, r *http.Request) {
	//

}
