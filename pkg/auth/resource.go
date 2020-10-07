package auth

import (
	"net/http"
	"io/ioutil"
	"errors"
	"encoding/json"

	"fmt"
	bson "go.mongodb.org/mongo-driver/bson"
)


type Resource struct {
	Id    string `json:"@id" bson:"@id"`
	Type  string `json:"@type" bson:"@type"`
	Owner string `json:"owner" bson:"owner"`
}

func (r Resource) ID() string {
	return r.Id
}

func (res Resource) Create() error {

	var err error

	res.Type = "Resource"

	// connect to the client
	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		return fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
	}

	// connect to collection
	collection := client.Database(MongoDatabase).Collection(MongoCollection)


	//TODO:  prove owner exists

	// create document
	_, err = collection.InsertOne(ctx, res)

	if err == nil {
		return nil
	}

	if ErrorDocumentExists(err) {
		return ErrDocumentExists
	}

	return err
}

func (r *Resource) Get() error {
	var b []byte
	var err error

	b, err = MongoFindOne(r.Id)
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

	err = res.Create()

	if err == nil {
		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"created": {"@id": "` + res.Id + `"}}`))
		return
	}

	// Error for when the User with u.Id Already Exists
	if errors.Is(err, ErrDocumentExists) {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"error": "User Already Exists" ,"@id": "` + res.Id + `"}`))
		return
	}

	// Unknown Error catch all
	w.WriteHeader(500)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write([]byte(`{"error": "` + err.Error() + `"}`))
	return

}

// TODO: (LowPriority) Write Handler for basic Get Resource by ID
func ResourceGet(w http.ResponseWriter, r *http.Request) {}

// TODO: (MidPriority) Write Handler for deletion by ID
func ResourceDelete(w http.ResponseWriter, r *http.Request) {}

// TODO: (LowPriority) Write Handler listing and filtering resources
func ResourceList(w http.ResponseWriter, r *http.Request) {}
