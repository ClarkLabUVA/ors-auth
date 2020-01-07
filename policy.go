package main

import (
	"net/http"
	"errors"
	"io/ioutil"
	"encoding/json"

	"fmt"
	bson "go.mongodb.org/mongo-driver/bson"
)

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


func PolicyCreate(w http.ResponseWriter, r *http.Request) {

	// read and marshal body json into
	var p Policy
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

	err = p.Create()

	if err == nil {
		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"created": {"@id": "` + p.Id + `"}}`))
		return
	}

	// Error for when the User with u.Id Already Exists
	if errors.Is(err, ErrDocumentExists) {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"error": "User Already Exists" ,"@id": "` + p.Id + `"}`))
		return
	}

	// Unknown Error catch all
	w.WriteHeader(500)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write([]byte(`{"error": "` + err.Error() + `"}`))
	return

}

// TODO: (LowPriority) Write Handler GetPolicy
func PolicyGet(w http.ResponseWriter, r *http.Request) {}

// TODO: (LowPriority) Write Handler PolicyUpdate
func PolicyUpdate(w http.ResponseWriter, r *http.Request) {}

// TODO: (LowPriority) Write Handler PolicyDelete
func PolicyDelete(w http.ResponseWriter, r *http.Request) {}

// TODO: (LowPriority) Write Handler PolicyList; filter by principal or resource as query params
func PolicyList(w http.ResponseWriter, r *http.Request) {}
