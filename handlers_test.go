package main

import (
  "testing"

  "encoding/json"
  "bytes"

  "net/http"
  "net/http/httptest"
)


func TestUserHandlers(t *testing.T) {
  u := User{
    Id: "orcid:1234-1234",
    Name: "Joe Schmoe",
    Email: "joe.schmoe@example.org",
    Admin: false,
    Groups: []string{},
  }

  t.Run("CreateUser", func(t *testing.T){

    userJSON, _ := json.Marshal(u)

    // user to JSON
    requestBody := bytes.NewReader(userJSON)

    // create a request to create a user
    request, _ := http.NewRequest("POST", "http://localhost:8080/user", requestBody)

    rr := httptest.NewRecorder()
    CreateUserHandler(rr, request)

    if rr.Code != 201 {
      t.Fatalf("Failed to Successfully Create User")
    }

  })

  t.Run("ListUsers", func(t *testing.T){


    // create a request to create a user
    request, _ := http.NewRequest("GET", "http://localhost:8080/user", nil)

    rr := httptest.NewRecorder()
    ListUserHandler(rr, request)

    if rr.Code != 200 {
      t.Fatalf("Failed To List Users Successfully")
    }

  })

  t.Run("GetUser", func(t *testing.T){

    // create a request to create a user
    request, _ := http.NewRequest("GET", "http://localhost:8080/user/"+u.Id, nil)

    rr := httptest.NewRecorder()
    GetUserHandler(rr, request)

    if rr.Code != 200 {
      t.Fatalf("Failed to Successfully Create User")
    }


  })

  t.Run("DeleteUser", func(t *testing.T){

    // create a request to create a user
    request, _ := http.NewRequest("DELETE", "http://localhost:8080/user/"+u.Id, nil)

    rr := httptest.NewRecorder()
    DeleteUserHandler(rr, request)

    if rr.Code != 200 {
      t.Fatalf("Failed to Successfully Create User")
    }
  })


}
