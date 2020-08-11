package auth

import (
	"errors"
	"net/http"
	"fmt"
	"io/ioutil"
	"strings"
	"encoding/base64"
	"encoding/json"
	"github.com/google/uuid"
)

var (
	errGlobusRevoke   = errors.New("Error Revoking Globus Token")
	errGlobusExchange = errors.New("Error Exchanging Globus Authorization Code for Globus Token")
	errHTTPInit       = errors.New("Error Creating HTTP Request")
	errHTTPRequest    = errors.New("Error Preforming HTTP Request")
)

// GlobusAuthClient is a struct for globus credentials and provides methods and handlers for the 3-legged oauth flow
type GlobusAuthClient struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	LogoutURL    string
	Scopes       string
}

func (g GlobusAuthClient) GrantHandler(w http.ResponseWriter, r *http.Request) {

	globusURL := "https://auth.globus.org/v2/oauth2/authorize?" +
		"scope=" + g.Scopes + "&" +
		"response_type=code" +
		"&client_id=" + g.ClientID +
		"&redirect_uri=" + g.RedirectURL +
		"&access_type=offline"

	http.Redirect(w, r, globusURL, 302)

}

func (g GlobusAuthClient) CodeHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	response := make(map[string]interface{})

	var code string
	code = r.FormValue("code")

	if code == "" {
		w.Write([]byte(`{"message": "URL missing requried query parameter", "error": "Missing Authorization Code"}`))
		w.WriteHeader(400)
		return
	}

	var err error
	var token GlobusAccessToken
	token, err = g.exchangeToken(code)

	// TODO: (LowPriority) Handle different errors for token exchange failing
	if err != nil {
		response["message"] = "failed to exchange authorization code with globus"
		response["error"] = err
		response["status_code"] = 400

		encodedResponse, _ := json.Marshal(response)
		w.Write(encodedResponse)
		w.WriteHeader(400)
		return
	}

	// introspect the token for user metadata
	var introspectedToken GlobusIntrospectedToken
	introspectedToken, err = g.introspectToken(token.AccessToken)
	if err != nil {
		response["message"] = "Failed to Introspect Globus Token"
		response["error"] = err
		response["status_code"] = 500
		response["globus_token"] = token

		encodedResponse, _ := json.Marshal(response)
		w.Write(encodedResponse)
		w.WriteHeader(500)
		return
	}

	// TODO: (LowestPriority) Add Support for Linked accounts identities
	//var identitiesResponse GlobusIdentitiesResponse
	//identitiesResponse, err = g.getIdentities( introspectedToken.IdentitiesSet )
	//if err != nil {
	// w.Write([]byte(`{"message": "Failed to Get Identities from Introspected Globus Token", "error": "`err.Error()`"}`) )
	//  w.WriteHeader(500)
	// return
	//}

	// find the user in the record
	user, err := queryUserEmail(introspectedToken.Email)

	/*
	// if no user record is found, create a new record
	if errors.Is(err, ErrNoDocument) {

		newUser, err := introspectedToken.registerUser()

		if err != nil {
			response["message"] = "Failed to register new user from token"
			response["error"] = err
			response["globus_token"] = introspectedToken

			encodedResponse, _ := json.Marshal(response)
			w.Write(encodedResponse)
			w.WriteHeader(500)
			return
		}

		response["globus_token"] = introspectedToken
		response["user"] = newUser

		encodedResponse, _ := json.Marshal(response)
		w.Write(encodedResponse)
		w.WriteHeader(201)
		return
	}
	*/

	// TODO: Return Error if Login isn't found
	// if error isn't no document found
	if err != nil {

		if err == errNoDocument {

			response["message"] = "No user record found"
			response["error"] = err.Error()
			response["status_code"] = 404
			response["globus_token"] = introspectedToken

			encodedResponse, _ := json.Marshal(response)
			w.Write(encodedResponse)
			w.WriteHeader(404)
			return

		}

		response["message"] = "Error Querying Database"
		response["error"] = err.Error()
		response["status_code"] = 500
		response["globus_token"] = introspectedToken

		encodedResponse, _ := json.Marshal(response)
		w.Write(encodedResponse)
		w.WriteHeader(500)
		return

	}

	response["user"] = user
	response["access_token"] = token
	response["introspected"] = introspectedToken
	response["status_code"] = 200
	//response["identities"] = identitiesResponse.Identities

	resp, err := json.Marshal(response)
	if err != nil {
		w.Write([]byte(`{"error": "error marshaling JSON", "error": "` + err.Error() + `"}`))
		w.WriteHeader(500)
		return
	}

	w.Write(resp)
	w.WriteHeader(200)

	return

}

// RevokeHandler is the http.HandlerFunc for revoking the globus token
func (g GlobusAuthClient) RevokeHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// get the bearer token
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	_, err := g.revokeToken(token)

	if err != nil {
		w.Write([]byte(`{"message": "Globus Token Revokation Failed", "error": "` + err.Error() + `"}`))
		w.WriteHeader(400)
		return
	}

	// responseBody, _ := json.Marshal(u)
	// w.Write(responseBody)
	w.Write([]byte(`{"message": "logged user out"}`))
	w.WriteHeader(200)
}

// InspectHandler is the http.HandlerFunc for inspecting tokens from globus and check membership
func (g GlobusAuthClient) InspectHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	var introspectedToken GlobusIntrospectedToken
	var user User
	var tok string
	var responseBody []byte

	// get bearer token from authorization header
	response := make(map[string]interface{})
	tok = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")


	// introspect globus token
	introspectedToken, err = g.introspectToken(tok)

	if err != nil {
		response["introspectedToken"] = introspectedToken
		response["error"] = err.Error()
		responseBody, _ = json.Marshal(response)
		w.Write(responseBody)
		w.WriteHeader(401)
		return
	}

	// check if user email is in users
	user, err = queryUserEmail(introspectedToken.Email)

	// if not return 404 user not found
	if err != nil {
		response["introspectedToken"] = introspectedToken
		response["error"] = err.Error()
		responseBody, _ = json.Marshal(response)
		w.Write(responseBody)
		w.WriteHeader(401)
		return
	}


	// else return 204 set response headers
	w.Header().Set("X-Client-ID", user.ID)
	w.WriteHeader(204)
	response["introspectedToken"] = introspectedToken
	response["user"] = user
	responseBody, _ = json.Marshal(response)
	w.Write(responseBody)
	return


}

// RefreshHandler is the http.HandlerFunc for granting a new access token from refresh tokens
// TODO: (MidPriority) Write Handler for Refresh Token Grant
func (g GlobusAuthClient) RefreshHandler(w http.ResponseWriter, r *http.Request) {}

func (g GlobusAuthClient) revokeToken(token string) (u User, err error) {

	client := &http.Client{}

	reqBody := strings.NewReader("token=" + token)

	req, err := http.NewRequest("POST", "https://auth.globus.org/v2/oauth2/token/revoke", reqBody)

	if err != nil {
		err = fmt.Errorf("%w: %s", errHTTPInit, err.Error())
		return
	}

	// set basic authentication header
	req.Header.Add(
		"Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(g.ClientID+":"+g.ClientSecret)),
	)

	// set content type header
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)

	// return error for issues with HTTP Request
	if err != nil {
		err = fmt.Errorf("%w: %s", errHTTPRequest, err.Error())
		return
	}

	// status code for successfull token revokation
	if resp.StatusCode != 200 {

		// if attempt to revoke token has failed
		responseBody, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("%w: %s", errGlobusRevoke, string(responseBody))
		return
	}

	//u, err = logoutUser(token)

	return

}

func (g GlobusAuthClient) exchangeToken(code string) (token GlobusAccessToken, err error) {

	client := &http.Client{}

	reqBody := strings.NewReader(
		"code=" + code +
			"&redirect_uri=" + g.RedirectURL +
			"&grant_type=authorization_code",
	)

	// exchange the authorization code for a globus access token
	req, err := http.NewRequest("POST", "https://auth.globus.org/v2/oauth2/token", reqBody)

	if err != nil {
		err = fmt.Errorf("%w: %s", errHTTPInit, err.Error())
		return
	}

	// encode basic authentication
	req.Header.Add(
		"Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(g.ClientID+":"+g.ClientSecret)),
	)

	// set content type header
	req.Header.Add(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	resp, err := client.Do(req)

	if err != nil {
		err = fmt.Errorf("%w: %s", errHTTPRequest, err.Error())
		return
	}

	respBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		// TODO: Fix Error Formatting for Failed Token Exchange
		// Error is produced when using an old token
		// err = fmt.Errorf("%w: %w", ErrGlobusExchange, string(respBody))
		err = errGlobusExchange
		return
	}

	err = json.Unmarshal(respBody, &token)

	if err != nil {
		err = fmt.Errorf("%w: %s", errJSONUnmarshal, err.Error())
		return
	}

	return

}

func (g GlobusAuthClient) introspectToken(token string) (introspectedToken GlobusIntrospectedToken, err error) {

	client := &http.Client{}

	reqBody := strings.NewReader("token=" + token + "&include=identities_set")

	req, err := http.NewRequest("POST", "https://auth.globus.org/v2/oauth2/token/introspect", reqBody)

	if err != nil {
		err = fmt.Errorf("%w: %s", errHTTPInit, err.Error())
		return
	}

	// encode basic authentication
	req.Header.Add(
		"Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(g.ClientID+":"+g.ClientSecret)),
	)

	req.Header.Add(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	resp, err := client.Do(req)

	if err != nil {
		err = fmt.Errorf("%w: %s", errHTTPRequest, err.Error())
		return
	}

	// if request to introspect token failed

	respBody, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(respBody, &introspectedToken)
	if err != nil {
		err = fmt.Errorf("%w: %s", errJSONUnmarshal, err.Error())
		return
	}

	return

}

func (g GlobusAuthClient) getIdentities(ids []string) (r GlobusIdentitiesResponse, err error) {

	client := &http.Client{}

	requestURI := "https://auth.globus.org/v2/api/identities?ids=" + strings.Join(ids, ",")
	req, err := http.NewRequest("GET", requestURI, nil)

	if err != nil {
		err = fmt.Errorf("%w: %s", errHTTPInit, err.Error())
		return
	}

	// encode basic authentication
	req.Header.Add(
		"Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(g.ClientID+":"+g.ClientSecret)),
	)

	resp, err := client.Do(req)

	if err != nil {
		err = fmt.Errorf("%w: %s", errHTTPRequest, err.Error())
		return
	}

	respBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		err = fmt.Errorf("%w: %s", errGlobusExchange, string(respBody))
		return
	}

	err = json.Unmarshal(respBody, &r)

	if err != nil {
		err = fmt.Errorf("%w: %s", errJSONUnmarshal, err.Error())
	}

	return

}

type GlobusAccessToken struct {
	AccessToken    string `json:"access_token"`
	Scope          string `json:"scope"`
	ResourceServer string `json:"resource_server"`
	ExpiresIn      int    `json:"expires_in"`
	TokenType      string `json:"token_type"`
	RefreshToken   string `json:"refresh_token"`
	State 				 string `json:"state"`
}

type GlobusIntrospectedToken struct {
	Active        bool     `json:"active"`
	Scope         string   `json:"scope"`
	Sub           string   `json:"sub"`
	Username      string   `json:"username"`
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	ClientID      string   `json:"client_id"`
	Audience      []string `json:"aud"`
	Expiration    int      `json:"exp"`
	IssuedAt      int      `json:"iat"`
	NotBefore     int      `json:"nbf"`
	IdentitiesSet []string `json:"identities_set"`
}

func (intro GlobusIntrospectedToken) findUser() (u User, err error) {
	u, err = queryUserEmail(intro.Email)

	if errors.Is(err, errMongoDecode) {
		return
	}

	return
}

func (intro GlobusIntrospectedToken) registerUser() (u User, err error) {

	userId, err := uuid.NewUUID()
	if err != nil {
		err = fmt.Errorf("%w: %w", errUUID, err)
		return
	}

	u.ID = userId.String()
	u.Name = intro.Name
	u.Email = intro.Email
	u.IsAdmin = false

	err = u.create()

	return
}

type GlobusIdentitiesResponse struct {
	Identities []GlobusIdentity `json:"identities"`
}

type GlobusIdentity struct {
	Username         string `json:"username"`
	Status           string `json:"status"`
	Name             string `json:"name"`
	Id               string `json:"id"`
	IdentityProvider string `json:"identity_provider"`
	Organization     string `json:"organization"`
	Email            string `json:"email"`
}
