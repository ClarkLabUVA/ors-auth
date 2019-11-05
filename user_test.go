package main


import (
  "testing"
)


func TestUser(t *testing.T) {

  MS := MongoServer{
    URI: "mongodb://mongoadmin:mongosecret@localhost:27017",
    Database: "auth",
    Collection: "auth",
    }

  var u = User{
    Id: "orcid:1234-1234-1234-1234",
    Name: "Joe Schmoe",
    Email: "JoeSchmoe@example.org",
    Admin: false,
    Groups: []string{},
  }

  t.Run("Create", func(t *testing.T){

    err := MS.CreateUser(u)

    if err != nil {
      t.Fatalf("Failed to Create the User: %s", err.Error())
    }

  })

  t.Run("Get", func(t *testing.T){

    foundUser, err := MS.GetUser(u.Id)
    if err != nil {
      t.Fatalf("Failed to Get User: %s", err.Error())
    }

    t.Logf("Found User: %+v", foundUser)

  })

  t.Run("List", func(t *testing.T){
    userList, err := MS.ListUser()

    if err != nil {
      t.Fatalf("Failed to List Users: %s", err.Error())
    }

    if len(userList) == 0 {
      t.Fatalf("Failed to List any Users")
    }

    t.Logf("Found Users: %+v", userList)

  })

  t.Run("Delete", func(t *testing.T){
    deletedUser, err := MS.DeleteUser(u.Id)

    if err != nil {
      t.Fatalf("Failed to Delete User: %s", err.Error())
    }

    t.Logf("Deleted User: %+v", deletedUser)

  })

}
