package main

import (
	"errors"
	"net/http"

	"fmt"
	"io/ioutil"
	"strings"

	"encoding/base64"
	"encoding/json"
)

var (
	ErrGlobusRevoke   = errors.New("Error Revoking Globus Token")
	ErrGlobusExchange = errors.New("Error Exchanging Globus Authorization Code for Globus Token")
	ErrHTTPInit       = errors.New("Error Creating HTTP Request")
	ErrHTTPRequest    = errors.New("Error Preforming HTTP Request")
)

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

	if err != nil {
		w.Write([]byte(`{"message": "failed to exchange authorization code with globus", "error": "` + err.Error() + `"}`))
		w.WriteHeader(400)
		return
	}

	// introspect the token for user metadata
	var introspectedToken GlobusIntrospectedToken
	introspectedToken, err = g.introspectToken(token.AccessToken)
	if err != nil {
		w.Write([]byte(`{"message": "Failed to Introspect Globus Token", "error": "` + err.Error() + `"}`))
		w.WriteHeader(500)
		return
	}

	// Ignoring Linked accounts for now

	//var identitiesResponse GlobusIdentitiesResponse
	//identitiesResponse, err = g.getIdentities( introspectedToken.IdentitiesSet )
	//if err != nil {
	// w.Write([]byte(`{"message": "Failed to Get Identities from Introspected Globus Token", "error": "`err.Error()`"}`) )
	//  w.WriteHeader(500)
	// return
	//}

	// find the user in the record
	user, err := queryUserEmail(introspectedToken.Email)

	if err != nil {
		w.Write([]byte(`{"email": "`+ introspectedToken.Email +`", "error": "User not registered"}`))
		w.WriteHeader(400)
		return
	}

	response := make(map[string]interface{})
	response["user"] = user
	response["access_token"] = token
	response["introspected"] = introspectedToken
	//response["identities"] = identitiesResponse.Identities

	resp, err := json.Marshal(response)
	if err != nil {
		w.Write([]byte(`{"error": "error marshaling JSON", "error": "` + err.Error() + `"}`))
		w.WriteHeader(500)
		return
	}

	w.Write(resp)
	w.Header().Set("ContentType", "application/json")
	w.WriteHeader(200)

	return

}

func (g GlobusAuthClient) RevokeHandler(w http.ResponseWriter, r *http.Request) {

	// get the bearer token
	token := strings.TrimPrefix("Bearer ", r.Header.Get("Authorization"))
	err := g.revokeToken(token)

	if err != nil {
		w.Write([]byte(`{"message": "Globus Token Revokation Failed", "error": "` + err.Error() + `"}`))
		w.WriteHeader(400)
		return
	}

	w.Header().Set("ContentType", "application/json")
	w.Write([]byte(`{"revoked": "` + token + `"}`))
	w.WriteHeader(200)
}

//func (g GlobusAuthClient) RefreshHandler(w http.ResponseWriter, r *http.Request) {}

//func (g GlobusAuthClient) RegisterHandler(w http.ResponseWriter, r *http.Request) {}

func (g GlobusAuthClient) revokeToken(token string) (err error) {

	client := &http.Client{}

	reqBody := strings.NewReader("token=" + token)

	req, err := http.NewRequest("POST", "https://auth.globus.org/v2/oauth2/token/revoke", reqBody)

	if err != nil {
		return fmt.Errorf("%w: %s", ErrHTTPInit, err.Error())
	}

	// set basic authentication header
	req.Header.Add(
		"Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(g.ClientID+":"+g.ClientSecret)),
	)

	// set content type header
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("%w: %s", ErrHTTPRequest, err.Error())
	}

	// status code for successfull token revokation
	if resp.StatusCode == 200 {
		return nil
	}

	// if attempt to revoke token has failed
	responseJson, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("%w: %s", ErrJSONUnmarshal, err.Error())
	}

	return fmt.Errorf("%w: %s", ErrGlobusRevoke, string(responseJson))

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
		err = fmt.Errorf("%w: %s", ErrHTTPInit, err.Error())
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
		err = fmt.Errorf("%w: %s", ErrHTTPRequest, err.Error())
		return
	}

	respBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {

		err = fmt.Errorf("%w: %s", ErrGlobusExchange, string(respBody))
		return
	}

	err = json.Unmarshal(respBody, &token)

	if err != nil {
		err = fmt.Errorf("%w: %s", ErrJSONUnmarshal, err.Error())
		return
	}

	return

}

func (g GlobusAuthClient) introspectToken(token string) (introspectedToken GlobusIntrospectedToken, err error) {

	client := &http.Client{}

	reqBody := strings.NewReader("token=" + token + "&include=identities_set")

	req, err := http.NewRequest("POST", "https://auth.globus.org/v2/oauth2/token/introspect", reqBody)

	if err != nil {
		err = fmt.Errorf("%w: %s", ErrHTTPInit, err.Error())
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
		err = fmt.Errorf("%w: %s", ErrHTTPRequest, err.Error())
		return
	}

	// if request to introspect token failed

	respBody, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(respBody, &introspectedToken)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrJSONUnmarshal, err.Error())
		return
	}

	return

}

func (g GlobusAuthClient) getIdentities(ids []string) (r GlobusIdentitiesResponse, err error) {

	client := &http.Client{}

	requestURI := "https://auth.globus.org/v2/api/identities?ids=" + strings.Join(ids, ",")
	req, err := http.NewRequest("GET", requestURI, nil)

	if err != nil {
		err = fmt.Errorf("%w: %s", ErrHTTPInit, err.Error())
		return
	}

	// encode basic authentication
	req.Header.Add(
		"Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(g.ClientID+":"+g.ClientSecret)),
	)

	resp, err := client.Do(req)

	if err != nil {
		err = fmt.Errorf("%w: %s", ErrHTTPRequest, err.Error())
		return
	}

	respBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		err = fmt.Errorf("%w: %s", ErrGlobusExchange, string(respBody))
		return
	}

	err = json.Unmarshal(respBody, &r)

	if err != nil {
		err = fmt.Errorf("%w: %s", ErrJSONUnmarshal, err.Error())
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
	//OIDToken       OIDToken  `json:"id_token"`
	State string `json:"state"`
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

	if errors.Is(err, ErrMongoDecode) {
		return
	}

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
