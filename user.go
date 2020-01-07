package main

import (
	"net/http"
	"io/ioutil"
	"errors"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"fmt"
	"regexp"
	"bytes"
	"github.com/google/uuid"
	bson "go.mongodb.org/mongo-driver/bson"
)


type User struct {
	Id      string   `json:"@id" bson:"@id"`
	Type    string   `json:"@type" bson:"@type"`
	Name    string   `json:"name" bson:"name"`
	Email   string   `json:"email" bson:"email"`
	IsAdmin bool     `json:"is_admin" bson:"is_admin"`
	Groups  []string `json:"groups" bson:"groups"`
	Session string   `json:"session" bson:"session"`
}

func (u User) ID() string {
	return u.Id
}

func (u User) MarshalJSON() ([]byte, error) {

	var userBuf bytes.Buffer
	var err error

	// open quotes
	userBuf.WriteString(`{`)

	// write out user id as full http
	userBuf.WriteString(fmt.Sprintf(`"@id": "%suser/%s"`, ORSURI, u.Id))
	userBuf.WriteString(`,`)

	// write context
	userBuf.WriteString(`"@context": {"@base": "http://schema.org/"}`)
	userBuf.WriteString(`,`)

	// write name
	userBuf.WriteString(fmt.Sprintf(`"name": "%s"`, u.Name))
	userBuf.WriteString(`,`)

	// write email
	userBuf.WriteString(fmt.Sprintf(`"email": "%s"`, u.Email))
	userBuf.WriteString(`,`)

	// groups
	userBuf.WriteString(`"memberOf": [`)

	if len(u.Groups) != 0 {
		for i, g := range u.Groups {
			userBuf.WriteString(fmt.Sprintf(`"%sgroup/%s"`, ORSURI, g))

			if i != len(u.Groups)-1 {
				userBuf.WriteString(`, `)
			}
		}

	}
	userBuf.WriteString(`]`)

	// close quote
	userBuf.WriteString(`}`)

	out := userBuf.Bytes()
	return out, err
}

func (u *User) UnmarshalJSON(data []byte) error {
	var err error

	aux := struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		IsAdmin bool   `json:"is_admin"`
	}{}

	if err = json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("%w: %s", ErrJSONUnmarshal, err.Error())
	}

	// validate name
	if aux.Name == "" {
		return fmt.Errorf("%w: User missing Name", ErrModelMissingField)
	}

	// validate email
	if aux.Email == "" {
		return fmt.Errorf("%w: User missing Email", ErrModelMissingField)
	}

	matched, err := regexp.MatchString(`^[a-zA-Z0-9-_.]*@[a-zA-Z]*\.[a-zA-Z]*$`, aux.Email)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrRegex, err.Error())
	}

	if !matched {
		return fmt.Errorf("%w: Invalid Email %s", ErrModelFieldValidation, aux.Email)
	}

	userId, err := uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrUUID, err.Error())
	}

	u.Id = userId.String()
	u.Type = TypeUser
	u.Name = aux.Name
	u.Email = aux.Email
	u.IsAdmin = aux.IsAdmin

	return err
}

func listUsers() (u []User, err error) {

	mongoCtx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
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

func (u *User) Create() (err error) {

	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
		return
	}

	// connect to the user collection
	collection := client.Database(MongoDatabase).Collection(MongoCollection)

	_, err = collection.InsertOne(ctx, u)

	// if errDocExists(err) {
	//	err = ErrDocumentExists
	// }

	return
}

func (u *User) Get() (err error) {
	var b []byte
	b, err = findOne(u.Id)
	if err != nil {
		return
	}

	err = bson.Unmarshal(b, &u)
	return
}

func queryUserEmail(email string) (u User, err error) {

	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
		return
	}

	// connect to the user collection
	collection := client.Database(MongoDatabase).Collection(MongoCollection)

	query := bson.D{{"email", email}, {"@type", TypeUser}}
	err = collection.FindOne(ctx, query).Decode(&u)

	// if decoding error
	if err != nil {
		return
	}

	return
}

func logoutUser(token string) (u User, err error) {

	query := bson.D{{"access_token", token}, {"@type", TypeUser}}
	update := bson.D{{"$set", bson.D{{"access_token", ""}}  }}

	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		return
	}

	collection := client.Database(MongoDatabase).Collection(MongoCollection)

	res := collection.FindOneAndUpdate(ctx, query, update)

	err = res.Decode(&u)

	// TODO: Error Handling For Following Cases
	// ErrNilDocument

	return

}

func (u *User) Delete() (err error) {
	var b []byte
	b, err = deleteOne(u.Id)
	if err != nil {
		return
	}

	err = bson.Unmarshal(b, &u)
	return
}

// TODO: (MidPriority) Add to User ListAccess()
// Return Everything a has adequate permissions to access
func (u User) ListAccess() (r []Resource, err error) {

	// query
	// query := bson.D{{""}}

	return
}

// TODO: (MidPriority) Add to User ListOwned()
func (u User) ListOwned() (r []Resource, err error) {
	return
}

// TODO: (MidPriority) Add to User ListPolicies()
// Return All Policies effecting this user
func (u User) ListPolicies() (p []Policy, err error) {

	return
}

// TODO: (MidPriority) Add to User ListChallenges()
// Return All Challenges this User has made
func (u User) ListChallenges() (c []Challenge, err error) {
	return
}

// POST /user/
func UserCreate(w http.ResponseWriter, r *http.Request) {
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
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"message": "Failed to Unmarshal Request JSON", "error": "%s"}`, err.Error() )
		return
	}

	err = u.Create()

	if err == nil {
		w.WriteHeader(201)
		fmt.Fprintf(w, `{"created": {"@id": "%s"}}`, u.Id)
		return
	}

	// Error for when the User with u.Id Already Exists
	if errors.Is(err, ErrDocumentExists) {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error": "User Already Exists" ,"@id": "%s"}`, u.Id)
		return
	}

	// Unknown Error catch all
	w.WriteHeader(500)
	fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
	return

}

// GET /user/
func UserList(w http.ResponseWriter, r *http.Request) {

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

// GET /user/:userID
func UserGet(w http.ResponseWriter, r *http.Request) {
	var u User
	var err error

	// get the user id from the route
	params := httprouter.ParamsFromContext(r.Context())

	u.Id = params.ByName("userID")
	err = u.Get()

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
func UserDelete(w http.ResponseWriter, r *http.Request) {

	var deletedUser User
	var err error

	// get the user id from the route
	params := httprouter.ParamsFromContext(r.Context())

	deletedUser.Id = params.ByName("userID")
	err = deletedUser.Delete()

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
