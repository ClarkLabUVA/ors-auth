package main

import (
	"flag"
	"github.com/julienschmidt/httprouter"
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
	router.Handler("POST", "/oauth/logout", http.HandlerFunc(globusClient.RevokeHandler))
	router.Handler("POST", "/oauth/refresh", http.HandlerFunc(globusClient.RefreshHandler))

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
