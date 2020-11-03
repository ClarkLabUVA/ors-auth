package auth

import (
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"errors"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"bytes"
	"github.com/google/uuid"
	bson "go.mongodb.org/mongo-driver/bson"

	"github.com/julienschmidt/httprouter"
)


// Group is a structure for group related data and operations
type Group struct {
	ID      string   `json:"@id" bson:"@id"`
	Type    string   `json:"@type" bson:"@type"`
	Name    string   `json:"name" bson:"name"`
	Admin   string   `json:"admin" bson:"admin"`
	Members []string `json:"members" bson:"members"`
}


// MarshalJSON is the custom method for converting the struct into JSON bytes
func (g Group) MarshalJSON() ([]byte, error) {

	var groupBuf bytes.Buffer
	var err error

	// open quotes
	groupBuf.WriteString(`{`)

	// write context
	groupBuf.WriteString(`"@context": {"@base": "http://schema.org/"}, "@type": "Organization", `)

	// write out user id as full http
	groupBuf.WriteString(fmt.Sprintf(`"@id": "%sgroup/%s", `, orsURI, g.ID))

	// write name
	groupBuf.WriteString(fmt.Sprintf(`"name": "%s", `, g.Name))

	// write admin
	groupBuf.WriteString(fmt.Sprintf(`"admin": "%s", `, g.Admin))

	// write members
	groupBuf.WriteString(`"member": [`)

	for i, mem := range g.Members {
		groupBuf.WriteString(fmt.Sprintf(`"%suser/%s"`,orsURI, mem))

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


// UnmarshalJSON is the custom method for converting the JSON bytes into the struct
func (g *Group) UnmarshalJSON(data []byte) error {
	var err error

	aux := struct {
		Name    string   `json:"name"`
		Admin   string   `json:"admin"`
		Members []string `json:"members"`
	}{}

	err = json.Unmarshal(data, &aux)
	if err != nil {
		return fmt.Errorf("%w: %s", errJSONUnmarshal, err.Error())
	}

	groupID, err := uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("%w: %s", errUUID, err.Error())
	}

	// validate name
	if aux.Name == "" {
		return fmt.Errorf("%w: Group missing Name", errModelMissingField)
	}

	// validate admin is not null
	if aux.Admin == "" {
		return fmt.Errorf("%w: Group missing Admin", errModelMissingField)
	}

	// add admin to members
	aux.Members = append(aux.Members, aux.Admin)

	g.ID = groupID.String()
	g.Name = aux.Name
	g.Admin = aux.Admin
	g.Members = aux.Members
	g.Type = typeGroup

	return err
}

func listGroups() (g []Group, err error) {

	mongoCtx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", errMongoClient, err.Error())
		return
	}

	collection := client.Database(mongoDatabase).Collection(mongoCollection)

	query := bson.D{{"@type", typeGroup}}
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

func (g *Group) get() (err error) {

	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", errMongoClient, err.Error())
		return
	}

	collection := client.Database(mongoDatabase).Collection(mongoCollection)
	err = collection.FindOne(ctx, bson.D{{"@id", g.ID}}).Decode(&g)

	return
}

// add owner to members
// update all members User profile to have
func (g *Group) create() (err error) {

	// connect to the client
	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", errMongoClient, err.Error())
		return
	}

	// connect to collection
	collection := client.Database(mongoDatabase).Collection(mongoCollection)

	//TODO: check that group doesn't already exist

	//TODO: check that Group Admin exists

	//TODO: error for users who don't exists

	// add membership of group
	adminResult := collection.FindOneAndUpdate(ctx,
		bson.D{{"@id", g.Admin}},
		bson.D{{"$addToSet", bson.D{{"groups", g.ID}}}},
	)

	if err = adminResult.Err(); err != nil {
		
		// TODO: error wrapping
		if  err == mongo.ErrNilDocument {
			err = fmt.Errorf("Unable to find Admin of Group: %w", err)
		}

		return
	}

	// Update each member to add Groups
	_, err = collection.UpdateMany(ctx,
		// elem match
		bson.D{{
			"@id", 
			bson.D{{
				"$in", 
				g.Members,
				}},
			}},
		bson.D{{
			"$addToSet", 
			bson.D{{
				"groups", 
				g.ID,
			}},
		}},
	)

	if err != nil {

		// TODO: error wrapping
		if  err == mongo.ErrNilDocument {
			err = fmt.Errorf("Unable to find member of Group: %w", err)
		}

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
func (g *Group) delete() error {
	var err error

	// connect to the client
	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		return fmt.Errorf("%w: %s", errMongoClient, err.Error())
	}

	// connect to collection
	collection := client.Database(mongoDatabase).Collection(mongoCollection)

	// Query for the Group, prove it exists
	err = collection.FindOne(ctx, bson.D{{"@id", g.ID}}).Decode(&g)
	if err != nil {
		return fmt.Errorf("DeleteGroupError: Group Not Found: %w", err)
	}

	// remove group from admin user
	_, err = collection.UpdateMany(ctx,
		bson.D{{
			"@id", 
			g.Admin,
		}},
		bson.D{{
			"$pull", 
			bson.D{{
				"groups", 
				g.ID,
			}},
		}},
	)

	// remove all users as members
	_, err = collection.UpdateMany(ctx,
		bson.D{{"@id", bson.D{{"$in", g.Members}}}},
		bson.D{{"$pull", bson.D{{"groups", g.ID}}}},
	)

	if err != nil {
		return fmt.Errorf("DeleteGroupError: Failed Updating Members: %w", err)
	}

	_, err = collection.DeleteOne(ctx,
		bson.D{{"@id", g.ID}, {"@type", typeGroup}},
	)

	if err != nil {
		return fmt.Errorf("DeleteGroupError: Failed Deleting Group: %w", err)
	}

	return nil
}

// TODO:
// User.Groups $addToSet groups to user
// Group.Members $addToSet
func (g Group) addUsers() {}

// TODO: Remove Users from a group
// User.Groups $pull
// Group.Members $pull
func (g Group) removeUsers() {}


// GroupCreate is the http handler for creating a single group
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

	err = g.create()

	if err == nil {
		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/ld+json")
		responseBody, _ := json.Marshal(g)
		w.Write(responseBody)
		return
	}

	// Error for when the User with u.Id Already Exists
	if errors.Is(err, errDocumentExists) {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"error": "User Already Exists" ,"@id": "` + g.ID + `"}`))
		return
	}

	// Unknown Error catch all
	w.WriteHeader(500)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write([]byte(`{"error": "` + err.Error() + `"}`))
	return

}


// GroupGet is the http handler for retrieving a single group by ID
func GroupGet(w http.ResponseWriter, r *http.Request) {

	var group Group
	var err error

	// get the user id from the route
	params := httprouter.ParamsFromContext(r.Context())

	group.ID = params.ByName("groupID")
	err = group.get()

	if err != nil {
		// TODO: error handling for p.get() 
		return
	}

	responseBytes, err := json.Marshal(group)

	if err != nil {
		// TODO: error handling for json marshal
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write(responseBytes)
	return

}


// GroupDelete is the http handler for deleting a single group
func GroupDelete(w http.ResponseWriter, r *http.Request) {

	var group Group
	var err error

	// get the user id from the route
	params := httprouter.ParamsFromContext(r.Context())

	group.ID = params.ByName("groupID")
	err = group.delete()

	if err != nil {
		// TODO: error handling for p.get() 
		return
	}

	responseBytes, err := json.Marshal(group)

	if err != nil {
		// TODO: error handling for json marshal
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write(responseBytes)
	return

}

// TODO
// GroupList is the http handler for listing all groups
func GroupList(w http.ResponseWriter, r *http.Request) {}

// TODO
// GroupUpdate is the http handler for updating a group
func GroupUpdate(w http.ResponseWriter, r *http.Request) {}
