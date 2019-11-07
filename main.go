package main

import (
  "errors"
  "net/http"
  "encoding/json"
  "io/ioutil"
  "github.com/julienschmidt/httprouter"
)

var MS = MongoServer{
  URI: "mongodb://mongoadmin:mongosecret@localhost:27017",
  Database: "auth",
  Collection: "auth",
  }


func main() {

  router := httprouter.New()

  router.Handler("POST", "/user", http.HandlerFunc(CreateUserHandler))
  router.Handler("GET", "/user", http.HandlerFunc(ListUserHandler))
  router.Handler("GET", "/user/:userID", http.HandlerFunc(GetUserHandler))
  router.Handler("DELETE", "/user/:userID", http.HandlerFunc(DeleteUserHandler))

  http.ListenAndServe(":8080", router)

}


// POST /user/
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {

  // read and marshal body json into
  var u User
  var err error
  var requestBody []byte
  requestBody, err = ioutil.ReadAll(r.Body)


  // If Error for Unmarshaling JSON Body
  if !json.Valid(requestBody) {
    w.WriteHeader(400)
    w.Header().Set("Content-Type", "application/ld+json")
    w.Write([]byte(`{"error": "Invalid JSON Submitted"}`))
    return
  }

  err = json.Unmarshal(requestBody, &u)

  if err != nil {
    w.WriteHeader(400)
    w.Header().Set("Content-Type", "application/ld+json")
    w.Write([]byte(`{"error": "Failed to Unmarshal Request JSON"}`))
    return
  }

  err = MS.CreateUser(u)

  if err == nil {
    w.WriteHeader(201)
    w.Header().Set("Content-Type", "application/ld+json")
    w.Write([]byte(`{"created": {"@id": "` + u.Id + `"}}`))
    return
  }

  // Error for when the User with u.Id Already Exists
  if errors.Is(err, ErrUserExists) {
    w.WriteHeader(400)
    w.Header().Set("Content-Type", "application/ld+json")
    w.Write([]byte(`{"error": "User Already Exists" ,"@id": "` + u.Id + `"}`))
    return
  }

  // Unknown Error catch all
  w.WriteHeader(500)
  w.Header().Set("Content-Type", "application/ld+json")
  w.Write([]byte(`{"error": "` + err.Error() + `"}`))
  return

}


// GET /user/
func ListUserHandler(w http.ResponseWriter, r *http.Request) {

  var err error
  var userList []User

  userList, err = MS.ListUser()

  if err != nil {
    return
  }

  var responseBody []byte
  responseBody, err = json.Marshal(userList)

  if err != nil {
      return
  }


  w.WriteHeader(200)
  w.Header().Set("Content-Type", "application/ld+json")
  w.Write(responseBody)

  return
}


// GET /user/:userID
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
  // get the user id from the route
  params := httprouter.ParamsFromContext(r.Context())
  userID := params.ByName("userID")

  u, err := MS.GetUser(userID)

  if err != nil {
    return
  }

  responseBytes, err := json.Marshal(u)

  if err != nil {
    return
  }

  w.WriteHeader(200)
  w.Header().Set("Content-Type", "application/ld+json")
  w.Write(responseBytes)
  return

}


// Delete /user/:userID
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {

  // get the user id from the route
  params := httprouter.ParamsFromContext(r.Context())
  userID := params.ByName("userID")

  deletedUser, err := MS.DeleteUser(userID)

  if err != nil {
    return
  }

  responseBytes, err := json.Marshal(deletedUser)
  if err != nil {
    return
  }

  w.WriteHeader(200)
  w.Header().Set("Content-Type", "application/ld+json")
  w.Write(responseBytes)
  return

}
