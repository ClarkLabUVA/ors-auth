package auth

import (
	"testing"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
)

func TestUserLogout(t *testing.T) {

	accessToken := "access"
	// setup test
	u := User{
		ID: "Test",
		Type: typeUser,
		Name: "TestUser",
		Email: "test@example.org",
		IsAdmin: false,
		AccessToken: accessToken,
		RefreshToken: "refresh",
	}

	err := u.create()

	if err != nil {
		t.Fatalf("Failed Setup: %s", err.Error())
	}


	t.Run("Success", func(t *testing.T){
		_, err := logoutUser(accessToken)

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
		found, err := logoutUser(accessToken)

		if err == nil {
			t.Errorf("Found User with Expired Token: %+v", found)
		}
	})

}

func TestUserHandlers(t *testing.T) {

	u := User{
		ID:      "orcid:1234-1234",
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
		UserCreateHandler(rr, request)

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
		UserListHandler(rr, request)

		if rr.Code != 200 {
			t.Fatalf("Failed To List Users Successfully")
		}

	})

	t.Run("GetUser", func(t *testing.T) {

		// create a request to create a user
		request := httptest.NewRequest("GET", "http://localhost:8080/user/"+u.ID, nil)

		rr := httptest.NewRecorder()
		UserGetHandler(rr, request)

		if rr.Code != 200 {
			t.Fatalf("Failed to Successfully Create User")
		}

	})

}

func TestUserMethods(t *testing.T) {

	var TestUser User

	TestUser.ID = "orcid:1234-1234-1234-1234"

	TestUser = User{
		ID:      "orcid:1234-1234-1234-1234",
		Type: 	typeUser,
		Name:    "Joe Schmoe",
		Email:   "JoeSchmoe@example.org",
		IsAdmin: false,
		Groups:  []string{},
	}

	t.Run("Create", func(t *testing.T) {

		err := TestUser.create()

		if err != nil {
			t.Errorf("Failed to Create the User: %s", err.Error())
		}

	})

	t.Run("Get", func(t *testing.T) {

		findUser := User{ID: TestUser.ID}
		err := findUser.get()

		if err != nil {
			t.Errorf("Failed to Get User: %s", err.Error())
		}

		t.Logf("Found User: %+v", findUser)

	})

	t.Run("GetEmail", func(t *testing.T) {
		u, err := queryUserEmail(TestUser.Email)
		if err != nil {
			t.Errorf("QueryUserEmail: Failed to Find TestUser Error: %s", err.Error())
		}

		t.Logf("QueryUserEmail: Found Test User \n\t%+v", u)

	})

	t.Run("List", func(t *testing.T) {
		userList, err := listUsers()

		if err != nil {
			t.Errorf("Failed to List Users: %s", err.Error())
		}

		if len(userList) == 0 {
			t.Errorf("Failed to List any Users")
		}

		t.Logf("Found Users: %+v", userList)

	})

}

func TestUserJSONUnmarshal(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		userBytes := []byte(`{"name": "Joe Schmoe", "email": "jschmoe@example.org", "is_admin": false}`)
		var u User

		err := json.Unmarshal(userBytes, &u)
		if err != nil {
			t.Fatalf("Error Unmarshaling Identifier")
		}

		t.Logf("UnmarshaledUser: %+v", u)

	})

	t.Run("InvalidEmail", func(t *testing.T) {

		userBytes := []byte(`{"name": "Joe Schmoe", "email": "jschmexample.org", "is_admin": false}`)
		var u User
		err := json.Unmarshal(userBytes, &u)
		if err == nil {
			t.Fatalf("ERROR: InvalidEmail \temail: %s", u.Email)
		}

		userBytes = []byte(`{"name": "Joe Schmoe", "email": "jschmexampleorg", "is_admin": false}`)
		err = json.Unmarshal(userBytes, &u)
		if err == nil {
			t.Fatalf("ERROR: %s", err.Error())
		}

		userBytes = []byte(`{"name": "Joe Schmoe", "email": "jschm@exampleorg", "is_admin": false}`)
		err = json.Unmarshal(userBytes, &u)
		if err == nil {
			t.Fatalf("ERROR: InvalidEmail \temail: %s", u.Email)
		}

		userBytes = []byte(`{"name": "Joe Schmoe", "email": "jschm@@example..org", "is_admin": false}`)
		err = json.Unmarshal(userBytes, &u)
		if err == nil {
			t.Fatalf("ERROR: InvalidEmail \temail: %s", u.Email)
		}

	})

	t.Run("ExtraFields", func(t *testing.T) {
		userBytes := []byte(`{"name": "Joe Schmoe", "email": "jschmoe@example.org", "is_admin": false, "groups": ["g1", "g2"]}`)
		var u User

		err := json.Unmarshal(userBytes, &u)
		if err != nil {
			t.Fatalf("Error Unmarshaling Identifier")
		}

		if len(u.Groups) != 0 {
			t.Fatalf("ErrGroups Not Empty:  %+v", u.Groups)
		}

	})
}

func TestUserJSONMarshal(t *testing.T) {

	u := User{
		ID:      "max",
		Email:   "mlev@example.org",
		Name:    "maxwell",
		Groups:  []string{"LevinsonFam", "Bagel Enthusiast"},
		IsAdmin: false,
		Session: "abcd",
	}
	userJSON, err := json.Marshal(u)

	if err != nil {
		t.Fatalf("ERROR: %s", err.Error())
	}

	t.Logf("MarshaledUser: %s", string(userJSON))

}
