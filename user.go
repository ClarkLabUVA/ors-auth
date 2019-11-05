package main

import (
  mongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

  "context"
  "time"

  bson "go.mongodb.org/mongo-driver/bson"

)




type MongoServer struct {
  URI string
  Database string
  Collection string
}


func  (ms *MongoServer) connect() (ctx context.Context, cancel context.CancelFunc, client *mongo.Client, err error) {

	// establish connection with mongo backend
	client, err = mongo.NewClient(options.Client().ApplyURI(ms.URI))
	if err != nil {
		return
	}

	// create a context for the connection
	ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)

	// connect to the client
	err = client.Connect(ctx)
	return
}



type User struct {
  Id            string      `json:"@id" bson:"_id"`
  Type          string      `json:"@type" bson:"@type"`
  Name          string      `json:"name" bson:"name"`
  Email         string      `json:"email" bson:"email"`
  Admin         bool        `json:"admin" bson:"admin"`
  Groups        []string    `json:"groups" bson:"groups"`
  //Session       string      `json:"session, omitempty"`
}


// TODO: validate orcid with regex
func validateORCID(orcid string) {}


// TODO: validate email with regex
func validateEmail(email string) {}


// List all Users available in the collection
func (ms *MongoServer)ListUser() (userList []User, err error) {

  mongoCtx, cancel, client, err := ms.connect()
  defer cancel()

  if err != nil {
    return
  }

  collection := client.Database(ms.Database).Collection(ms.Collection)

  cur, err := collection.Find(mongoCtx, bson.D{{"@type", "Person"}}, nil)
  defer cur.Close(mongoCtx)
  if err != nil {
    return
  }

  err = cur.All(mongoCtx, &userList)
  if err != nil {
    return
  }

  return

}


// Create a New User with Data within the Struct
func (ms *MongoServer)CreateUser(u User) (err error) {

  ctx, cancel, client, err := ms.connect()
  defer cancel()

  if err != nil { return }

  // connect to the user collection
  collection := client.Database(ms.Database).Collection(ms.Collection)

  // default values
  u.Type = "Person"
  bsonRecord, err := bson.Marshal(u)
  if err != nil {
    return
  }

  _, err = collection.InsertOne(ctx, bsonRecord)
  if err != nil { return }

  return
}


// Query User data by userId
func (ms *MongoServer)GetUser(userId string) (u User, err error) {

  ctx, cancel, client, err := ms.connect()
  defer cancel()

  if err != nil { return }

  // connect to the user collection
  collection := client.Database(ms.Database).Collection(ms.Collection)

  err = collection.FindOne(ctx, bson.D{{"_id", userId }}).Decode(&u)

  return
}


// Delete the User specified by Id
func (ms *MongoServer)DeleteUser(userId string) (u User, err error) {

  ctx, cancel, client, err := ms.connect()
  defer cancel()

  if err != nil { return }

  // connect to the user collection
  collection := client.Database(ms.Database).Collection(ms.Collection)


  res := collection.FindOneAndDelete(ctx, bson.D{{"_id", userId}})
  err = res.Decode(&u)

  return

}
