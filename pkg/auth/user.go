package auth

import (
	"net/http"
	"io/ioutil"
	"strings"
	"errors"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"log"
	"fmt"
	"github.com/google/uuid"
	bson "go.mongodb.org/mongo-driver/bson"
	mongo "go.mongodb.org/mongo-driver/mongo"
	options "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/dgrijalva/jwt-go"
    "time"
    "os"
)

var jwtSecret []byte

func init() {

    jwtENV, ok := os.LookupEnv("JWT_SECRET")

    if !ok {
        jwtENV = "test secret"
    }

    jwtSecret = []byte(jwtENV)
}

func createUserIndex() {

	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		log.Printf("UserInit: Failed to Connect to Mongo\t Error: %s", err.Error())
	}

	collection := client.Database(mongoDatabase).Collection(mongoCollection)

	// build an index creation request
	model := mongo.IndexModel{
		Keys: bson.D{{"email", 1}},
		Options: options.Index().SetName("email").SetUnique(true).SetPartialFilterExpression(bson.D{{"@type",  typeUser}}),
	}

	opts := options.CreateIndexes()


	// create unique index on email property for documents of type
	_, err = collection.Indexes().CreateOne(ctx, model, opts)

	if err != nil {
		log.Printf("Setting Up User Index: %s", err.Error())
	}

}

// User is a structure for user data with methods for interacting with Mongo
type User struct {
	ID      string   `json:"@id" bson:"@id"`
	Type    string   `json:"@type" bson:"@type"`
	Name    string   `json:"name" bson:"name"`
	Email   string   `json:"email" bson:"email"`
	Role    string   `json:"role" bson:"role"`
	Groups  []string `json:"groups" bson:"groups"`
	AccessToken  string  `json:"access_token" bson:"access_token"`
	RefreshToken string	`json:"refresh_token" bson:"refresh_token"`
}

type UserTokenClaims struct {
	Role string `json:"role"`
	Groups string `json:"groups"`
	jwt.StandardClaims
}

// loginUser creates a JWT for the user for use with ORS services
// it also stores this token within the user record for easy verification
func (u *User) newSession() (err error) {

	now := time.Now()
	
	claims := UserTokenClaims {
		u.Role,
		strings.Join(u.Groups, ";"),
		jwt.StandardClaims{
			Subject: u.ID,
			Audience: "https://fairscape.org",
			IssuedAt: now.Unix(),
			ExpiresAt: now.Add(time.Hour * 48).Unix(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    u.AccessToken, err = accessToken.SignedString(jwtSecret)

    return
}

// parseToken takes the string value of the token and returns the user information contained within
func parseToken(tokenString string) (u User, err error) {

	token, err := jwt.ParseWithClaims(tokenString, &UserTokenClaims{}, func(token *jwt.Token) (interface {}, error) {
		// no secret currently
		return jwtSecret, nil
	})


	claims := token.Claims.(*UserTokenClaims)

	log.Printf("%+v", claims)

	return 
}


func queryUserEmail(email string) (u User, err error) {

	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", errMongoClient, err.Error())
		return
	}

	// connect to the user collection
	collection := client.Database(mongoDatabase).Collection(mongoCollection)

	query := bson.D{{"email", email}, {"@type", typeUser}}
	err = collection.FindOne(ctx, query).Decode(&u)

	// TODO: Error Handling
	// if decoding error
	if err == mongo.ErrNoDocuments {
		err = errNoDocument
	}

	return
}


func logoutUser(token string) (u User, err error) {

	query := bson.D{{"access_token", token}, {"@type", typeUser}}
	update := bson.D{{"$set", bson.D{{"access_token", ""}}  }}

	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		return
	}

	collection := client.Database(mongoDatabase).Collection(mongoCollection)

	res := collection.FindOneAndUpdate(ctx, query, update)

	err = res.Decode(&u)

	// TODO: Error Handling For Following Cases
	// ErrNilDocument

	return

}


func listUsers() (u []User, err error) {

	mongoCtx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", errMongoClient, err.Error())
		return
	}

	collection := client.Database(mongoDatabase).Collection(mongoCollection)

	query := bson.D{{"@type", typeUser}}
	cur, err := collection.Find(mongoCtx, query, nil)
	defer cur.Close(mongoCtx)

	if err != nil {
		err = fmt.Errorf("%w: %s", errMongoQuery, err.Error())
		return
	}

	err = cur.All(mongoCtx, &u)
	if err != nil {
		return
	}

	return

}


func (u *User) create() (err error) {

	uid, err := uuid.NewRandom()
	u.ID = uid.String()

	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", errMongoClient, err.Error())
		return
	}

	// connect to the user collection
	collection := client.Database(mongoDatabase).Collection(mongoCollection)

	_, err = collection.InsertOne(ctx, u)

	if err == nil {
		return
	}

	if errorDocumentExists(err) {
		err = errDocumentExists
	}

	return
}


func (u *User) get() (err error) {

	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", errMongoClient, err.Error())
		return
	}

	collection := client.Database(mongoDatabase).Collection(mongoCollection)
	err = collection.FindOne(ctx, bson.D{{"@id", u.ID}}).Decode(&u)

	return
}

func (u *User) updateToken() (err error) {

	ctx, cancel, client, err := connectMongo()
	defer cancel()

	// connect to the user collection
	collection := client.Database(mongoDatabase).Collection(mongoCollection)

    filter := bson.D{{"@id", u.ID}}
    update := bson.D{{"$set", bson.D{{"access_token", u.AccessToken}} }}

	_, err = collection.UpdateOne(ctx, filter, update)

    return
}


func (u *User) delete() (err error) {

	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", errMongoClient, err.Error())
		return
	}

	collection := client.Database(mongoDatabase).Collection(mongoCollection)
	err = collection.FindOneAndDelete(ctx, bson.D{{"@id", u.ID}}).Decode(&u)

	return
}

// TODO: (MidPriority) Add to User ListAccess()
// Return Everything a has adequate permissions to access
/*
func (u User) listAccess() (r []Resource, err error) {

	// query
	// query := bson.D{{""}}

	return
}

// TODO: (MidPriority) Add to User ListOwned()
func (u User) listOwned() (r []Resource, err error) {
	return
}

// TODO: (MidPriority) Add to User ListPolicies()
// Return All Policies effecting this user
func (u User) listPolicies() (p []Policy, err error) {

	return
}

// TODO: (MidPriority) Add to User ListChallenges()
// Return All Challenges this User has made
func (u User) ListChallenges() (c []Challenge, err error) {
	return
}
*/

// UserCreateHandler is the handler for creating a User
// POST /user/ with JSON body
func UserCreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/ld+json")

	// read and marshal body json into
	var u User
	var err error
	var requestBody []byte

	requestBody, err = ioutil.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"message": "Failed to " ,"error": "%s"}`, err.Error() )
		return
	}

	// If Error for Unmarshaling JSON Body
	if !json.Valid(requestBody) {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error": "Invalid JSON Submitted"}`)
		return
	}

	err = json.Unmarshal(requestBody, &u)

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, `{"message": "Failed to Unmarshal Request JSON", "error": "%s"}`, err.Error() )
		return
	}

	err = u.create()

	if err == nil {
		w.WriteHeader(201)
		createdUser, _ := json.Marshal(u)
		w.Write(createdUser)
		return
	}

	// Error for when the User with u.Id Already Exists
	if errors.Is(err, errDocumentExists) {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error": "User Already Exists" ,"@id": "%s"}`, u.ID)
		return
	}

	// Unknown Error catch all
	w.WriteHeader(500)
	fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
	return

}

// UserListHandler is the handler
// GET /user/
func UserListHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	var userList []User

	userList, err = listUsers()

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

// UserGetHandler gets the details for a single user
// GET /user/:userID
func UserGetHandler(w http.ResponseWriter, r *http.Request) {
	var u User
	var err error

	// get the user id from the route
	params := httprouter.ParamsFromContext(r.Context())

	u.ID = params.ByName("userID")
	err = u.get()

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

// User Delete Handler
// DELETE /user/:userID
func UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var u User
	var err error

	// get the user id from the route
	params := httprouter.ParamsFromContext(r.Context())

	u.ID = params.ByName("userID")
	err = u.delete()

	if err != nil {
		w.WriteHeader(500)
		return
	}

	responseBytes, err := json.Marshal(u)

	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write(responseBytes)
	return

}
