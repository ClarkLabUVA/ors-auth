package auth

import (
	"net/http"
	"errors"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"bytes"
	"github.com/google/uuid"
	bson "go.mongodb.org/mongo-driver/bson"
)

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

	query := bson.D{{"@type", TypeGroup}}
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
	b, err = MongoFindOne(g.Id)
	if err != nil {
		return
	}

	err = bson.Unmarshal(b, &g)
	return
}

// add owner to members
// update all members User profile to have
func (g *Group) Create() (err error) {

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
		bson.D{{"@id", g.Admin}, {"@type", TypeUser}},
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

func GroupCreate(w http.ResponseWriter, r *http.Request) {

	// read and marshal body json into
	var g Group
	var err error
	var requestBody []byte

	requestBody, err = ioutil.ReadAll(r.Body)

	// If Error for Unmarshaling JSON Body
	if !json.Valid(requestBody) {
		w.WriteHeader(400)
		w.Write([]byte(`{"error": "Invalid JSON Submitted"}`))
		return
	}

	err = json.Unmarshal(requestBody, &r)

	if err != nil {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"error": "Failed to Unmarshal Request JSON"}`))
		return
	}

	err = g.Create()

	if err == nil {
		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/ld+json")
		responseBody, _ := json.Marshal(g)
		w.Write(responseBody)
		return
	}

	// Error for when the User with u.Id Already Exists
	if errors.Is(err, ErrDocumentExists) {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"error": "User Already Exists" ,"@id": "` + g.Id + `"}`))
		return
	}

	// Unknown Error catch all
	w.WriteHeader(500)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write([]byte(`{"error": "` + err.Error() + `"}`))
	return

}

// TODO: (LowPriority) Write Endpoint GroupGet
func GroupGet(w http.ResponseWriter, r *http.Request) {}

// TODO: (LowPriority) Write Endpoint GroupUpdate
func GroupUpdate(w http.ResponseWriter, r *http.Request) {}

// TODO: (LowPriority) Write Endpoint GroupDelete
func GroupDelete(w http.ResponseWriter, r *http.Request) {}

// TODO: (LowPriority) Write Endpoint GroupList
func GroupList(w http.ResponseWriter, r *http.Request) {}
