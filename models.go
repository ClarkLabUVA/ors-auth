package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	//"strings"
	"encoding/json"
	"reflect"
	"regexp"
	"time"

	"github.com/google/uuid"

	bson "go.mongodb.org/mongo-driver/bson"
	mongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrNoDocument						= errors.New("NoDocumentFound")
	ErrDocumentExists       = errors.New("DocumentExists")
	ErrMongoClient          = errors.New("MongoClientError")
	ErrMongoQuery           = errors.New("MongoQueryError")
	ErrMongoDecode          = errors.New("MongoDecodeError")
	ErrModelFieldValidation = errors.New("ErrorModelFieldValidation")
	ErrModelMissingField    = errors.New("ErrorModelMissingRequiredField")
	ErrJSONUnmarshal        = errors.New("ErrorParsingJSON")
	ErrUUID                 = errors.New("ErrorCreatingUUID")
	ErrRegex                = errors.New("ErrorRunningRegex")
)

var (
	MongoURI        = "mongodb://mongoadmin:mongosecret@localhost:27017"
	MongoDatabase   = "test"
	MongoCollection = "auth"
)

var (
	ORSURI = "http://ors.uvadcos.io/"
)

const (
	TypeGroup     = "Group"
	TypeUser      = "Person"
	TypeResource  = "Resource"
	TypePolicy    = "Policy"
	TypeChallenge = "Challenge"
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

	if errDocExists(err) {
		err = ErrDocumentExists
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
		err = fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
		return
	}

	// connect to the user collection
	collection := client.Database(MongoDatabase).Collection(MongoCollection)

	query := bson.D{{"@id", Id}}
	b, err = collection.FindOneAndDelete(ctx, query).DecodeBytes()

	return
}

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
	userBuf.WriteString(fmt.Sprintf(`"name": "%s"`, u.Email))
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

func (u User) Create() (err error) {

	u.Type = TypeUser

	err = insertOne(u)

	if errDocExists(err) {
		return ErrDocumentExists
	}

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


type Group struct {
	Id      string   `json:"@id" bson:"@id"`
	Type    string   `json:"@type" bson:"@type"`
	Name    string   `json:"name" bson:"name"`
	Admin   string   `json:"admin" bson:"admin"`
	Members []string `json:"members" bson:"members"`
}

func (g Group) ID() string {
	return g.Id
}

func (g Group) MarshalJSON() ([]byte, error) {

	var groupBuf bytes.Buffer
	var err error

	// open quotes
	groupBuf.WriteString(`{`)

	// write context
	groupBuf.WriteString(`"@context": {"@base": "http://schema.org/"}, "@type": "Organization", `)

	// write out user id as full http
	groupBuf.WriteString(fmt.Sprintf(`"@id": "%sgroup/%s", `, ORSURI, g.Id))

	// write name
	groupBuf.WriteString(fmt.Sprintf(`"name": "%s", `, g.Name))

	// write admin
	groupBuf.WriteString(fmt.Sprintf(`"admin": "%s", `, g.Admin))

	// write members
	groupBuf.WriteString(`"member": [`)

	for i, mem := range g.Members {
		groupBuf.WriteString(fmt.Sprintf(`"%suser/%s"`, ORSURI, mem))

		if i != len(g.Members)-1 {
			groupBuf.WriteString(`, `)
		}
	}
	groupBuf.WriteString(`]`)

	// close quote
	groupBuf.WriteString(`}`)

	out := groupBuf.Bytes()
	return out, err
}

func (g *Group) UnmarshalJSON(data []byte) error {
	var err error

	aux := struct {
		Name    string   `json:"name"`
		Admin   string   `json:"admin"`
		Members []string `json:"members"`
	}{}

	err = json.Unmarshal(data, &aux)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrJSONUnmarshal, err.Error())
	}

	groupId, err := uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrUUID, err.Error())
	}

	// validate name
	if aux.Name == "" {
		return fmt.Errorf("%w: Group missing Name", ErrModelMissingField)
	}

	// validate admin is not null
	if aux.Admin == "" {
		return fmt.Errorf("%w: Group missing Admin", ErrModelMissingField)
	}

	// add admin to members
	aux.Members = append(aux.Members, aux.Admin)

	g.Id = groupId.String()
	g.Name = aux.Name
	g.Admin = aux.Admin
	g.Members = aux.Members
	g.Type = TypeGroup

	return err
}

func listGroups() (g []Group, err error) {

	mongoCtx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
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

func (g *Group) Get() (err error) {
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
func (g Group) Create() (err error) {

	// connect to the client
	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
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
		bson.D{{"$addToSet", bson.D{{"groups", g.Id}}}},
	)

	if err = adminResult.Err(); err != nil {
		// TODO: error wrapping
		return
	}

	// Update each member to add Groups
	_, err = collection.UpdateMany(ctx,
		// elem match
		bson.D{{"@id", bson.D{{"$elemMatch", bson.D{{"@id", g.Members}}}}}},
		bson.D{{"$addToSet", bson.D{{"groups", g.Id}}}},
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
func (g *Group) Delete() error {
	var err error

	// connect to the client
	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		return fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
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
		bson.D{{"@id", bson.D{{"$in", g.Members}}}},
		bson.D{{"$pull", bson.D{{"groups", g.Id}}}},
	)

	if err != nil {
		return fmt.Errorf("DeleteGroupError: Failed Updating Members: %w", err)
	}

	// remove from policies
	_, err = collection.UpdateMany(ctx,
		bson.D{{"@type", "Policy"}},
		bson.D{{"$pull", bson.D{{"principal", g.Id}}}},
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
func (g Group) AddUsers() {}

// TODO: Remove Users from a group
// User.Groups $pull
// Group.Members $pull
func (g Group) RemoveUsers() {}

type Resource struct {
	Id    string `json:"@id" bson:"@id"`
	Type  string `json:"@type" bson:"@type"`
	Owner string `json:"owner" bson:"owner"`
}

func (r Resource) ID() string {
	return r.Id
}

func (r Resource) Create() error {

	// prove owner exists
	r.Type = "Resource"

	err := insertOne(r)

	if errDocExists(err) {
		return ErrDocumentExists
	}

	return err
}

func (r *Resource) Get() error {
	var b []byte
	var err error

	b, err = findOne(r.Id)
	if err != nil {
		return err
	}

	err = bson.Unmarshal(b, &r)
	return err

}

func (r *Resource) Delete() error {

	var err error

	// connect to the client
	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		return fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
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
		err = fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
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
	Id        string   `json:"@id" bson:"@id"`
	Type      string   `json:"@type" bson:"@type"`
	Resource  string   `json:"resource" bson:"resource"`
	Principal []string `json:"principal" bson:"principal"`
	Effect    string   `json:"effect" bson:"effect"`
	Action    []string `json:"action" bson:"action"`
	Issuer    string   `json:"issuer" bson:"issuer"`
}

func (p Policy) ID() string {
	return p.Id
}

func listPolicies() (p []Policy, err error) {

	mongoCtx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("MongoDB: %w: %s", ErrMongoClient, err.Error())
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

func (p Policy) Create() error {

	var err error

	// connect to the client
	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		return fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
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

func (p *Policy) Get() error {

	var b []byte
	var err error

	b, err = findOne(p.Id)
	if err != nil {
		return err
	}

	err = bson.Unmarshal(b, &p)
	return err

}

func (p *Policy) Delete() error {

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

func (c Challenge) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	var err error

	// open quotes
	buf.WriteString(`{`)

	// write context
	buf.WriteString(`"@context": {"@base": "http://schema.org/", "principal": "http://schema.org/agent"}, "@type": "AuthorizeAction", `)

	// write out challenge id as full url
	buf.WriteString(fmt.Sprintf(`"@id": "%schallenge/%s", `, ORSURI, c.Id))

	// write principal
	buf.WriteString(fmt.Sprintf(`"principal": "%s", `, c.Principal))

	// write resource
	buf.WriteString(fmt.Sprintf(`"resource": "%s", `, c.Resource))

	// write action
	buf.WriteString(fmt.Sprintf(`"action": "%s", `, c.Action))

	// write time
	jsonTime, _ := c.Time.MarshalJSON()
	buf.WriteString(fmt.Sprintf(`"time": %s, `, string(jsonTime)))

	// write granted
	buf.WriteString(fmt.Sprintf(`"granted": %t }`, c.Granted))

	out := buf.Bytes()
	return out, err
}

func (c *Challenge) UnmarshalJSON(data []byte) error {
	var err error

	type Alias Challenge
	aux := struct {
		Type    string    `json:"@type" bson:"@type"`
		Time    time.Time `json:"time" bson:"time"`
		Granted bool      `json:"granted" bson:"granted"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	err = json.Unmarshal(data, &aux)

	if err != nil {
		return err
	}

	challengeId, err := uuid.NewUUID()

	if err != nil {
		return err
	}

	c.Id = challengeId.String()

	c.Type = TypeChallenge
	c.Time = time.Now()
	c.Granted = false

	return nil
}

func (c Challenge) ID() string {
	return c.Id
}

func (c *Challenge) Evaluate() (err error) {

	c.Type = "Challenge"
	c.Time = time.Now()
	c.Granted = false

	// connect to the client
	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		return fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
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
					bson.D{{"$in", []string{"*", c.Action}}}}}},
				{"principal", bson.D{{"$elemMatch",
					bson.D{{"$in", append(u.Groups, u.Id)}}}}},
			})

		if err != nil {
			return
		}

		err = cursor.All(ctx, &matchingPolicies)

		if err != nil {
			return
		}

		if len(matchingPolicies) == 0 {
			c.Granted = false
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
		err = fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
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

// Function to determine if an error from the mongo server is a MongoWriteError
// where the document already exists and collides on at least one unique index
func errDocExists(err error) bool {

	// if the mongo operation returned a Write Exception
	if errorType := reflect.TypeOf(err); errorType.Name() == "WriteException" {

		writeErr := err.(mongo.WriteException).WriteErrors

		if writeErr[0].Code == 11000 {
			return true
		}
	}
	return false
}
