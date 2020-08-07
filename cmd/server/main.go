package main

import (
	"log"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/fairscape/auth/pkg/auth"
	"os"
)

func main() {


	var globusClientID = os.Getenv("GLOBUS_CLIENT_ID")
	var globusClientSecret = os.Getenv("GLOBUS_CLIENT_SECRET")
	var redirectURL = os.Getenv("GLOBUS_REDIRECT_URL")

	var scopes = "urn:globus:auth:scope:auth.globus.org:view_identity_set+urn:globus:auth:scope:auth.globus.org:view_identities+openid+email+profile"

	globusClient := auth.GlobusAuthClient{
		ClientID:     globusClientID,
		ClientSecret: globusClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
	}

	router := httprouter.New()

	// oauth token routes
	router.Handler("POST", "/inspect", http.HandlerFunc(globusClient.InspectHandler))
	router.Handler("GET", "/token", http.HandlerFunc(globusClient.CodeHandler))
	router.Handler("POST", "/refresh", http.HandlerFunc(globusClient.RefreshHandler))
	router.Handler("GET", "/login", http.HandlerFunc(globusClient.GrantHandler))
	router.Handler("POST", "/logout", http.HandlerFunc(globusClient.RevokeHandler))

	// user managment routes
	router.Handler("POST", "/user", http.HandlerFunc(auth.UserCreateHandler))
	router.Handler("GET", "/user", http.HandlerFunc(auth.UserListHandler))
	router.Handler("GET", "/user/:userID", http.HandlerFunc(auth.UserGetHandler))
	router.Handler("DELETE", "/user/:userID", http.HandlerFunc(auth.UserDeleteHandler))

	/*
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

	log.Fatal(http.ListenAndServe(":80", router))

}
