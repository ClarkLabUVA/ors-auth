package main

import (
	"testing"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
)

func TestUserLogout(t *testing.T) {

	access_token := "access"
	// setup test
	u := User{
		Id: "Test",
		Type: TypeUser,
		Name: "TestUser",
		Email: "test@example.org",
		IsAdmin: false,
		AccessToken: access_token,
		RefreshToken: "refresh",
	}

	err := u.Create()

	if err != nil {
		t.Fatalf("Failed Setup: %s", err.Error())
	}


	t.Run("Success", func(t *testing.T){
		_, err := logoutUser(access_token)

		if err != nil {
			t.Errorf("Failed Logout: %s", err.Error())
		}

		t.Logf("Successfully Logged Out")

	})

	t.Run("NoUser", func(t *testing.T) {
		found, err := logoutUser("fakeToken")

		if err == nil {
			t.Errorf("Found User for Fake Token: %+v", found)
		}
	})

	t.Run("AlreadyLoggedOut", func(t *testing.T){
		found, err := logoutUser(access_token)

		if err == nil {
			t.Errorf("Found User with Expired Token: %+v", found)
		}
	})

	// tear down tests
	u.Delete()

}

func TestUserHandlers(t *testing.T) {

	u := User{
		Id:      "orcid:1234-1234",
		Name:    "Joe Schmoe",
		Email:   "joe.schmoe@example.org",
		IsAdmin: false,
		Groups:  []string{},
	}

	t.Run("CreateUser", func(t *testing.T) {

		userJSON, err := json.Marshal(u)
		if err != nil {
			t.Errorf("Failed Marshaling JSON: %s", err.Error())
		}

		// user to JSON
		requestBody := bytes.NewReader(userJSON)

		// create a request to create a user
		request := httptest.NewRequest("POST", "http://localhost:8080/user", requestBody)

		rr := httptest.NewRecorder()
		UserCreate(rr, request)

		resp := rr.Result()
		body, _ := ioutil.ReadAll(resp.Body)


		if resp.StatusCode != 201 {
			t.Errorf("StatusCode: %d \nBody: %s", resp.StatusCode, string(body))
		}

	})

	t.Run("ListUsers", func(t *testing.T) {

		// create a request to create a user
		request := httptest.NewRequest("GET", "http://localhost:8080/user", nil)

		rr := httptest.NewRecorder()
		UserList(rr, request)

		if rr.Code != 200 {
			t.Fatalf("Failed To List Users Successfully")
		}

	})

	t.Run("GetUser", func(t *testing.T) {

		// create a request to create a user
		request := httptest.NewRequest("GET", "http://localhost:8080/user/"+u.Id, nil)

		rr := httptest.NewRecorder()
		UserGet(rr, request)

		if rr.Code != 200 {
			t.Fatalf("Failed to Successfully Create User")
		}

	})

	t.Run("DeleteUser", func(t *testing.T) {

		// create a request to create a user
		request := httptest.NewRequest("DELETE", "http://localhost:8080/user/"+u.Id, nil)

		rr := httptest.NewRecorder()
		UserDelete(rr, request)

		if rr.Code != 200 {
			t.Fatalf("Failed to Successfully Create User")
		}
	})

}
