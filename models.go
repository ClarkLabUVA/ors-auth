package main

import (
  "fmt"
  "errors"
  "reflect"
  "context"
  "time"

  mongo "go.mongodb.org/mongo-driver/mongo"
  bson "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
  ErrDocumentExists = errors.New("DocumentExists")
  ErrMongoClient= errors.New("MongoClientError")
  ErrMongoQuery = errors.New("MongoQueryError")
  ErrMongoDecode = errors.New("MongoDecodeError")
)


var (
  MongoURI = "mongodb://mongoadmin:mongosecret@localhost:27017"
  MongoDatabase = "test"
  MongoCollection = "auth"
)

func connectMongo() (ctx context.Context, cancel context.CancelFunc, client *mongo.Client, err error) {

	// establish connection with mongo backend
	client, err = mongo.NewClient(options.Client().ApplyURI(MongoURI))
	if err != nil {
		return
	}

	// create a context for the connection
	ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)

	// connect to the client
	err = client.Connect(ctx)
	return
}


func insertOne(document interface{}) (err error) {
  ctx, cancel, client, err := connectMongo()
  defer cancel()

  if err != nil {
    err = fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
    return
  }

  // connect to the user collection
  collection := client.Database(MongoDatabase).Collection(MongoCollection)

  _, err = collection.InsertOne(ctx, document)
  if err != nil {
    if errDocExists(err) {
      // rewrite error message
      //err = fmt.Errorf("%w: %s", ErrUserExists, query["@id"])
      err = ErrDocumentExists
      return
    }
  }


  return
}


func findOne(Id string) (b []byte, err error) {
  ctx, cancel, client, err := connectMongo()
  defer cancel()

  if err != nil {
    err = fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
    return
  }

  // connect to the user collection
  collection := client.Database(MongoDatabase).Collection(MongoCollection)

  query := bson.D{{"@id", Id}}
  b, err = collection.FindOne(ctx, query).DecodeBytes()

  if err != nil {
    return
  }

  return
}


func deleteOne(Id string) (b []byte, err error) {
  ctx, cancel, client, err := connectMongo()
  defer cancel()

  if err != nil {
    err = fmt.Errorf("MongoDB: %w: %s", ErrClientFailure, err.Error())
    return
  }

  // connect to the user collection
  collection := client.Database(MongoDatabase).Collection(MongoCollection)

  query := bson.D{{"@id", Id}}
  b, err = collection.FindOneAndDelete(ctx, query).DecodeBytes()

  return
}


type User struct {
  Id            string      `json:"@id" bson:"@id"`
  Type          string      `json:"@type" bson:"@type"`
  Name          string      `json:"name" bson:"name"`
  Email         string      `json:"email" bson:"email"`
  IsAdmin       bool        `json:"is_admin" bson:"is_admin"`
  Groups        []string    `json:"groups" bson:"groups"`
  Session       string      `json:"session, omitempty" bson:"session"`
}


func listUsers() (u []User, err error) {

  mongoCtx, cancel, client, err := connectMongo()
  defer cancel()

  if err != nil {
    err = fmt.Errorf("MongoDB: %w: %s", ErrClientFailure, err.Error())
    return
  }


  collection := client.Database(MongoDatabase).Collection(MongoCollection)


  query := bson.D{{"@type", "User"}}
  cur, err := collection.Find(mongoCtx, query, nil)
  defer cur.Close(mongoCtx)
  if err != nil {
    err = fmt.Errorf("%w: %s", ErrMongoQuery, err.Error())
    return
  }

  err = cur.All(mongoCtx, &u)
  if err != nil {
    return
  }

  return


}



func (u User)Create() (err error) {

  u.Type = "User"

  err = insertOne(u)
  return
}



func (u *User)Get() (err error) {
  var b []byte
  b, err = findOne(u.Id)
  if err != nil {
    return
  }

  err = bson.Unmarshal(b, &u)
  return
}



func (u *User)Delete() (err error) {
  var b []byte
  b, err = deleteOne(u.Id)
  if err != nil {
    return
  }

  err = bson.Unmarshal(b, &u)
  return
}



// Return Everything a has adequate permissions to access
func (u User)ListAccess() (r []Resource, err error) {

  // query
  // query := bson.D{{""}}


  return
}


// Return all Resources
func (u User)ListOwned() (r []Resource, err error) {
  return
}



// Return All Policies effecting this user
func (u User)ListPolicies() (p []Policy, err error) {

  return
}


// Return All Challenges this User has made
func (u User)ListChallenges() (c []Challenge, err error) {
  return
}



type Group struct {
  Id      string   `json:"@id" bson:"@id"`
  Type    string   `json:"@type" bson:"@type"`
  Name    string   `json:"name" bson:"name"`
  Admin   string   `json:"admin" bson:"admin"`
  Members []string `json:"members" bson:"members"`
}


func listGroups() (g []Group, err error) {

  mongoCtx, cancel, client, err := connectMongo()
  defer cancel()

  if err != nil {
    err = fmt.Errorf("MongoDB: %w: %s", ErrClientFailure, err.Error())
    return
  }

  collection := client.Database(MongoDatabase).Collection(MongoCollection)

  query := bson.D{{"@type", "Group"}}
  cur, err := collection.Find(mongoCtx, query, nil)
  defer cur.Close(mongoCtx)
  if err != nil {
    return
  }

  err = cur.All(mongoCtx, &g)
  if err != nil {
    return
  }

  return


}



func (g *Group)Get() (err error) {
  var b []byte
  b, err = findOne(g.Id)
  if err != nil {
    return
  }

  err = bson.Unmarshal(b, &g)
  return
}



// add owner to members
// update all members User profile to have
func (g Group)Create() (err error) {

  // connect to the client
  ctx, cancel, client, err := connectMongo()
  defer cancel()

  if err != nil {
    err = fmt.Errorf("MongoDB: %w: %s", ErrClientFailure, err.Error())
    return
  }

  // connect to collection
  collection := client.Database(MongoDatabase).Collection(MongoCollection)

  //TODO: check that group doesn't already exist

  //TODO: check that Group Admin exists

  //TODO: error for users who don't exists


  // add membership of group
  adminResult := collection.FindOneAndUpdate(ctx,
      bson.D{{"@id", g.Admin}, {"@type", "User"}},
      bson.D{{"$addToSet", bson.D{{"groups", g.Id}} }},
  )

  if err = adminResult.Err(); err != nil {
    // TODO: error wrapping
    return
  }


  // Update each member to add Groups
  _, err = collection.UpdateMany(ctx,
    bson.D{{"@id", bson.D{{"$in", g.Members}} }},
    bson.D{{"$addToSet", bson.D{{"groups", g.Id}} }},
  )

  if err != nil {
    // TODO: error wrapping

    // TODO: undo admin update
    return
  }

  _, err = collection.InsertOne(ctx, g)
  if err != nil {

    // TODO: error wrapping

    // TODO: undo admin update

    // TODO: undo members update

    return
  }

  return

}



// remove all users as being members
// remove from all policies as a principal
//  -if group is the last principal on the Policy, delete the policy
// delete group entry
func (g *Group)Delete() (error) {
  var err error

  // connect to the client
  ctx, cancel, client, err := connectMongo()
  defer cancel()

  if err != nil {
    return fmt.Errorf("DeleteGroupError: %w: %s", ErrClientFailure, err.Error())
  }

  // connect to collection
  collection := client.Database(MongoDatabase).Collection(MongoCollection)

  // Query for the Group, prove it exists
  err = collection.FindOne(ctx, bson.D{{"@id", g.Id}}).Decode(&g)
  if err != nil {
    return fmt.Errorf("DeleteGroupError: Group Not Found: %w", err)
  }

  // remove all users as members
  _, err = collection.UpdateMany(ctx,
    bson.D{{ "@id",   bson.D{{ "$in", g.Members}} }},
    bson.D{{ "$pull", bson.D{{ "groups", g.Id}} }},
  )

  if err != nil {
    return fmt.Errorf("DeleteGroupError: Failed Updating Members: %w", err)
  }

  // remove from policies
  _, err = collection.UpdateMany(ctx,
    bson.D{{ "@type", "Policy"}},
    bson.D{{ "$pull", bson.D{{ "principal", g.Id}} }},
  )

  if err != nil {
    return fmt.Errorf("DeleteGroupError: Failed Updating Policies: %w", err)
  }

  _, err = collection.DeleteOne(ctx,
    bson.D{{"@id", g.Id}, {"@type", "Group"}},
  )

  if err != nil {
    return fmt.Errorf("DeleteGroupError: Failed Deleting Group: %w", err)
  }


  return nil
}



// TODO:
// User.Groups $addToSet groups to user
// Group.Members $addToSet
func (g Group)AddUsers() {}



// TODO: Remove Users from a group
// User.Groups $pull
// Group.Members $pull
func (g Group)RemoveUsers() {}





type Resource struct {
  Id       string `json:"@id" bson:"@id"`
  Type     string `json:"@type" bson:"@type"`
  Owner    string `json:"owner" bson:"owner"`
}



func (r Resource)Create() (error) {

  // prove owner exists

  r.Type = "Resource"
  return insertOne(r)
}



func (r *Resource)Get() (error) {
  var b []byte
  var err error

  b, err = findOne(r.Id)
  if err != nil {
    return err
  }

  err = bson.Unmarshal(b, &r)
  return err

}



func (r *Resource)Delete() (error) {

  var err error

  // connect to the client
  ctx, cancel, client, err := connectMongo()
  defer cancel()

  if err != nil {
    return fmt.Errorf("MongoDB: %w: %s", ErrClientFailure, err.Error())
  }

  // connect to collection
  collection := client.Database(MongoDatabase).Collection(MongoCollection)

  // Query for the Resource, prove it exists
  err = collection.FindOne(ctx, bson.D{{"@id", r.Id}}).Decode(&r)
  if err != nil {
    return fmt.Errorf("DeleteResourceError: Group Not Found: %w", err)
  }

  // Delete All Policies with Resource
  _, err = collection.DeleteMany(ctx,
    bson.D{{"resource", r.Id}, {"@type", "Policy"}},
    )

  if err != nil {
    return fmt.Errorf("DeleteResourceError: Failed to Delete Policies: %w", err)
  }

  // Delete the Resource Object
  _, err = collection.DeleteOne(ctx, bson.D{{"@id", r.Id}, {"@type", "Resource"}})

  if err != nil {
    return fmt.Errorf("DeleteResourceError: Failed to Delete Resource: %w", err)
  }

  return nil
}



func listResources() (r []Resource, err error) {

  mongoCtx, cancel, client, err := connectMongo()
  defer cancel()

  if err != nil {
    err = fmt.Errorf("MongoDB: %w: %s", ErrClientFailure, err.Error())
    return
  }

  collection := client.Database(MongoDatabase).Collection(MongoCollection)

  query := bson.D{{"@type", "Resource"}}
  cur, err := collection.Find(mongoCtx, query, nil)
  defer cur.Close(mongoCtx)
  if err != nil {
    return
  }

  err = cur.All(mongoCtx, &r)
  if err != nil {
    return
  }

  return


}





type Policy struct {
  Id            string   `json:"@id" bson:"@id"`
  Type          string   `json:"@type" bson:"@type"`
  Resource      string   `json:"resource" bson:"resource"`
  Principal     []string `json:"principal" bson:"principal"`
  Effect        string   `json:"effect" bson:"effect"`
  Action        []string `json:"action" bson:"action"`
  Issuer        string   `json:"issuer" bson:"issuer"`
}



func listPolicies() (p []Policy, err error) {

  mongoCtx, cancel, client, err := connectMongo()
  defer cancel()

  if err != nil {
    err = fmt.Errorf("MongoDB: %w: %s", ErrClientFailure, err.Error())
    return
  }

  collection := client.Database(MongoDatabase).Collection(MongoCollection)

  query := bson.D{{"@type", "Policy"}}
  cur, err := collection.Find(mongoCtx, query, nil)
  defer cur.Close(mongoCtx)
  if err != nil {
    return
  }

  err = cur.All(mongoCtx, &p)
  if err != nil {
    return
  }

  return


}



func (p Policy)Create() (error) {

  var err error

  // connect to the client
  ctx, cancel, client, err := connectMongo()
  defer cancel()

  if err != nil {
    return fmt.Errorf("%w: %s", ErrClientFailure, err.Error())
  }

  // connect to collection
  collection := client.Database(MongoDatabase).Collection(MongoCollection)

  // prove resource exists
  var r Resource
  err = collection.FindOne(ctx, bson.D{{"@id", p.Resource}, {"@type", "Resource"}}).Decode(&r)

  if err != nil {
    return err
  }

  // TODO: prove all principals exist
  // cur, err = collection.Find(ctx, bson.D{{"@id", bson.D{{"$in", p.Principal}}   }})


  // create policy record
  p.Type = "Policy"
  _, err = collection.InsertOne(ctx, p)

  if err != nil {
    return err
  }

  return nil
}


func (p *Policy)Get() (error) {

  var b []byte
  var err error

  b, err = findOne(p.Id)
  if err != nil {
    return err
  }

  err = bson.Unmarshal(b, &p)
  return err

}


func (p *Policy)Delete() (error) {

  var b []byte
  var err error

  b, err = deleteOne(p.Id)
  if err != nil {
    return err
  }

  err = bson.Unmarshal(b, &p)
  return err

}



type Challenge struct {
  Id        string    `json:"@id" bson:"@id"`
  Type      string    `json:"@type" bson:"@type"`
  Principal string    `json:"principal" bson:"principal"`
  Resource  string    `json:"resource" bson:"resource"`
  Action    string    `json:"action" bson:"action"`
  Time      time.Time `json:"time" bson:"time"`
  Issuer    string    `json:"issuer" bson:"issuer"`
  Granted   bool      `json:"granted" bson:"granted"`
}



func (c *Challenge)Evaluate() (err error) {

  c.Type = "Challenge"
  c.Time = time.Now()
  c.Granted = false


  // connect to the client
  ctx, cancel, client, err := connectMongo()
  defer cancel()

  if err != nil {
    return fmt.Errorf("%w: %s", ErrClientFailure, err.Error())
  }

  // connect to collection
  collection := client.Database(MongoDatabase).Collection(MongoCollection)


  // get principal
  var u User
  err = collection.FindOne(ctx, bson.D{{"@id", c.Principal}}).Decode(&u)

  // get resource
  var r Resource
  err = collection.FindOne(ctx, bson.D{{"@id", c.Resource}}).Decode(&r)

  // if principal is owner
  if c.Principal != r.Owner {

    // search policies for granting relevant permissions
    var matchingPolicies []Policy

    // change to a regex match for the action
    // i.e. action s3:download can be matched by s3:*
    var cursor *mongo.Cursor
    cursor, err = collection.Find(ctx,
      bson.D{
        {"@type", "Policy"},
        {"effect", "Allow"},
        {"resource", c.Resource},
        {"action", bson.D{{"$elemMatch",
          bson.D{{"$in", []string{"*", c.Action} }} }} },
        {"principal", bson.D{{"$elemMatch",
          bson.D{{"$in", append(u.Groups, u.Id) }}  }} },
      })

    if err != nil {
      return
    }

    err = cursor.All(ctx, &matchingPolicies)

    if err != nil {
      return
    }

    if len(matchingPolicies) == 0 {
      c.Granted= false
    } else {
      c.Granted = true
    }

  } else {
    c.Granted = true
  }

  _, err = collection.InsertOne(ctx, c)
  return

}



func listChallenges() (c []Challenge, err error) {

  mongoCtx, cancel, client, err := connectMongo()
  defer cancel()

  if err != nil {
    err = fmt.Errorf("MongoDB: %w: %s", ErrClientFailure, err.Error())
    return
  }

  collection := client.Database(MongoDatabase).Collection(MongoCollection)

  query := bson.D{{"@type", "Challenge"}}
  cur, err := collection.Find(mongoCtx, query, nil)
  defer cur.Close(mongoCtx)
  if err != nil {
    return
  }

  err = cur.All(mongoCtx, &c)
  if err != nil {
    return
  }

  return


}



func errDocExists(err error) (bool) {

    // if the mongo operation returned a Write Exception
    if errorType := reflect.TypeOf(err); errorType.Name() == "WriteException" {

      writeErr := err.(mongo.WriteException).WriteErrors

      if writeErr[0].Code == 11000 {
        return true
      }
    }
    return false
}
