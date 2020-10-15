package auth

import (
	"net/http"
	"io/ioutil"
	"errors"
	"encoding/json"

	"fmt"
	bson "go.mongodb.org/mongo-driver/bson"
		
	"github.com/julienschmidt/httprouter"
)


// Resource is a structure for documenting entities contained in the framework
type Resource struct {
	ID    string `json:"@id" bson:"@id"`
	Type  string `json:"@type" bson:"@type"` 
	Owner string `json:"owner" bson:"owner"`
}

func (r *Resource) create() error {

	var err error

	r.Type = "Resource"

	// connect to the client
	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		return fmt.Errorf("%w: %s", errMongoClient, err.Error())
	}

	// connect to collection
	collection := client.Database(mongoDatabase).Collection(mongoCollection)


	//TODO:  prove owner exists

	// create document
	_, err = collection.InsertOne(ctx, r)

	if err == nil {
		return nil
	}

	if errorDocumentExists(err) {
		return errDocumentExists
	}

	return err
}

func (r *Resource) get() error {

	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", errMongoClient, err.Error())
		return err
	}

	collection := client.Database(mongoDatabase).Collection(mongoCollection)
	err = collection.FindOne(ctx, bson.D{{"@id", r.ID}}).Decode(&r)

	return err

}

func (r *Resource) delete() error {

	var err error

	// connect to the client
	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		return fmt.Errorf("%w: %s", errMongoClient, err.Error())
	}

	// connect to collection
	collection := client.Database(mongoDatabase).Collection(mongoCollection)

	// Query for the Resource, prove it exists
	err = collection.FindOne(ctx, bson.D{{"@id", r.ID}}).Decode(&r)
	if err != nil {
		return fmt.Errorf("DeleteResourceError: Group Not Found: %w", err)
	}

	// Delete All Policies with Resource
	_, err = collection.DeleteMany(ctx,
		bson.D{{"resource", r.ID}, {"@type", "Policy"}},
	)

	if err != nil {
		return fmt.Errorf("DeleteResourceError: Failed to Delete Policies: %w", err)
	}

	// Delete the Resource Object
	_, err = collection.DeleteOne(ctx, bson.D{{"@id", r.ID}, {"@type", "Resource"}})

	if err != nil {
		return fmt.Errorf("DeleteResourceError: Failed to Delete Resource: %w", err)
	}

	return nil
}

func listResources() (r []Resource, err error) {

	mongoCtx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", errMongoClient, err.Error())
		return
	}

	collection := client.Database(mongoDatabase).Collection(mongoCollection)

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


// ResourceCreate is the http handler for the api operation to create a resource
func ResourceCreate(w http.ResponseWriter, r *http.Request) {

	// read and marshal body json into
	var res Resource
	var err error
	var requestBody []byte

	requestBody, err = ioutil.ReadAll(r.Body)

	// If Error for Unmarshaling JSON Body
	if !json.Valid(requestBody) {
		w.WriteHeader(400)
		w.Write([]byte(`{"error": "Invalid JSON Submitted"}`))
		return
	}

	err = json.Unmarshal(requestBody, &res)

	if err != nil {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"error": "Failed to Unmarshal Request JSON"}`))
		return
	}

	err = res.create()

	if err == nil {
		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"created": {"@id": "` + res.ID + `"}}`))
		return
	}

	// Error for when the User with u.Id Already Exists
	if errors.Is(err, errDocumentExists) {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"error": "User Already Exists" ,"@id": "` + res.ID + `"}`))
		return
	}

	// Unknown Error catch all
	w.WriteHeader(500)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write([]byte(`{"error": "` + err.Error() + `"}`))
	return

}

// ResourceGet is the http handler for retrieving a single resource by ID
func ResourceGet(w http.ResponseWriter, r *http.Request) {

	
	var resource Resource
	var err error

	// get the user id from the route
	params := httprouter.ParamsFromContext(r.Context())

	resource.ID = params.ByName("resourceID")
	err = resource.get()

	if err != nil {
		// TODO: error handling for resource.get() 
		return
	}

	responseBytes, err := json.Marshal(resource)

	if err != nil {
		// TODO: error handling for json marshal
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write(responseBytes)
	return

}

// ResourceDelete is the http handler for deleting a single resource by ID
func ResourceDelete(w http.ResponseWriter, r *http.Request) {

	var resource Resource
	var err error

	// get the user id from the route
	params := httprouter.ParamsFromContext(r.Context())

	resource.ID = params.ByName("resourceID")
	err = resource.delete()

	if err != nil {
		// TODO: error handling for resource.get() 
		return
	}

	responseBytes, err := json.Marshal(resource)

	if err != nil {
		// TODO: error handling for json marshal
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write(responseBytes)
	return

}

// ResourceList is for listing all the current resources
func ResourceList(w http.ResponseWriter, r *http.Request) {}
