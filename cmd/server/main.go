//© 2020 By The Rector And Visitors Of The University Of Virginia

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package server 

import (
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"github.com/ClarkLabUVA/auth/pkg/globus"
	"os"
)

func main() {

	router := httprouter.New()

	var globusClientID = os.Getenv("GLOBUS_CLIENT_ID")
	var globusClientSecret = os.Getenv("GLOBUS_CLIENT_SECRET")
	var redirectURL = os.Getenv("REDIRECT_URL")
	
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
