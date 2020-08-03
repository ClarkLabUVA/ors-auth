package server 

import (
	"log"
	"net/http"
	"github.com/fairscape/auth/pkg/auth"
	"os"
)

func main() {


	var globusClientID = os.Getenv("GLOBUS_CLIENT_ID")
	var globusClientSecret = os.Getenv("GLOBUS_CLIENT_SECRET")
	var redirectURL = os.Getenv("REDIRECT_URL")
	
	var scopes = "urn:globus:auth:scope:auth.globus.org:view_identity_set+urn:globus:auth:scope:auth.globus.org:view_identities+openid+email+profile"

	globusClient := auth.GlobusAuthClient{
		ClientID:     globusClientID,
		ClientSecret: globusClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
	}

	router := http.NewServeMux()
	
	// any method
	router.HandleFunc("/oauth/inspect", globusClient.InspectHandler)

	// GET
	router.HandleFunc("/oauth/login", globusClient.GrantHandler)

	// GET /oauth/token
	// handle the code grant from 
	router.HandleFunc("/oauth/token", globusClient.CodeHandler)

	// POST /oauth/logout
	// invalidate a current session within globus auth and locally
	router.HandleFunc("/oauth/logout", globusClient.RevokeHandler)

	// POST /oauth/refresh  
	// using refresh token and grant a new access token
	router.HandleFunc("/oauth/refresh", globusClient.RefreshHandler)

	/*
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
	*/

	log.Fatal(http.ListenAndServe("0.0.0.0:8080", router))

}
