//© 2020 By The Rector And Visitors Of The University Of Virginia

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
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

    router.Handler("POST", "/resource", http.HandlerFunc(auth.ResourceCreate))
    router.Handler("GET", "/resource", http.HandlerFunc(auth.ResourceList))
    router.Handler("GET", "/resource/*resourceID", http.HandlerFunc(auth.ResourceGet))
    router.Handler("DELETE", "/resource/*resourceID", http.HandlerFunc(auth.ResourceDelete))

    router.Handler("POST", "/group", http.HandlerFunc(auth.GroupCreate))
    router.Handler("GET", "/group", http.HandlerFunc(auth.GroupList))
    router.Handler("GET", "/group/:groupID", http.HandlerFunc(auth.GroupGet))
    router.Handler("PUT", "/group/:groupID", http.HandlerFunc(auth.GroupUpdate))
	router.Handler("DELETE", "/group/:groupID", http.HandlerFunc(auth.GroupDelete))

	log.Fatal(http.ListenAndServe(":8080", router))

}
