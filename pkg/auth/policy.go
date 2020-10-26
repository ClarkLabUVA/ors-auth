package auth

import (
	"net/http"
	"errors"
	"io/ioutil"
	"encoding/json"

	"fmt"
	bson "go.mongodb.org/mongo-driver/bson"

	"github.com/julienschmidt/httprouter"
)

// Policy is a structure for policy data and associated operations
type Policy struct {
	ID        string   `json:"@id" bson:"@id"`
	Type      string   `json:"@type" bson:"@type"`
	Resource  string   `json:"resource" bson:"resource"`
	Principal []string `json:"principal" bson:"principal"`
	Effect    string   `json:"effect" bson:"effect"`
	Action    []string `json:"action" bson:"action"`
	Issuer    string   `json:"issuer" bson:"issuer"`
}

func listPolicies() (p []Policy, err error) {

	mongoCtx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("MongoDB: %w: %s", errMongoClient, err.Error())
		return
	}

	collection := client.Database(mongoDatabase).Collection(mongoCollection)

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

func (p Policy) create() error {

	var err error

	// connect to the client
	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		return fmt.Errorf("%w: %s", errMongoClient, err.Error())
	}

	// connect to collection
	collection := client.Database(mongoDatabase).Collection(mongoCollection)

	// prove resource exists
	var r Resource
	err = collection.FindOne(ctx, bson.D{{"@id", p.Resource}, {"@type", "Resource"}}).Decode(&r)

	if err != nil {
		return err
	}

	// TODO: prove all principals exist
	// cur, err = collection.Find(ctx, bson.D{{"@id", bson.D{{"$in", p.Principal}}   }})

	// create policy record
	p.Type = typePolicy
	_, err = collection.InsertOne(ctx, p)

	if err != nil {
		return err
	}

	return nil
}

func (p *Policy) get() error {

	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", errMongoClient, err.Error())
		return err
	}

	collection := client.Database(mongoDatabase).Collection(mongoCollection)
	err = collection.FindOne(ctx, bson.D{{"@id", p.ID}}).Decode(&p)

	return err

}

func (p *Policy) delete() error {

	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", errMongoClient, err.Error())
		return err
	}

	collection := client.Database(mongoDatabase).Collection(mongoCollection)
	err = collection.FindOneAndDelete(ctx, bson.D{{"@id", p.ID}}).Decode(&p)

	return err
}


// TODO
func (p *Policy) update() error { 
	return nil
}


// PolicyCreate is the http handler for the api opertaion for creating a policy
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

	err = json.Unmarshal(requestBody, &p)

	if err != nil {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"error": "Failed to Unmarshal Request JSON"}`))
		return
	}

	err = p.create()

	if err == nil {
		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/ld+json")
		responseBody, _ := json.Marshal(p)
		w.Write(responseBody)
		return
	}

	// Error for when the User with u.Id Already Exists
	if errors.Is(err, errDocumentExists) {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/ld+json")
		w.Write([]byte(`{"error": "User Already Exists" ,"@id": "` + p.ID + `"}`))
		return
	}

	// Unknown Error catch all
	w.WriteHeader(500)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write([]byte(`{"error": "` + err.Error() + `"}`))
	return

}


// PolicyGet is the http handler for the retrieval of a policy by the 
func PolicyGet(w http.ResponseWriter, r *http.Request) {

	var p Policy
	var err error

	// get the user id from the route
	params := httprouter.ParamsFromContext(r.Context())

	p.ID = params.ByName("policyID")
	err = p.get()

	if err != nil {
		// TODO: error handling for p.get() 
		return
	}

	responseBytes, err := json.Marshal(p)

	if err != nil {
		// TODO: error handling for json marshal
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write(responseBytes)
	return
}


// PolicyDelete is the http handler used for deleting a single policy by ID
func PolicyDelete(w http.ResponseWriter, r *http.Request) {

	var p Policy
	var err error

	// get the user id from the route
	params := httprouter.ParamsFromContext(r.Context())

	p.ID = params.ByName("policyID")
	err = p.delete()

	if err != nil {
		// TODO: error handling for p.get() 
		return
	}

	responseBytes, err := json.Marshal(p)

	if err != nil {
		// TODO: error handling for json marshal
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/ld+json")
	w.Write(responseBytes)
	return


}


// PolicyList is the http handler for listing all policies
func PolicyList(w http.ResponseWriter, r *http.Request) {}



// PolicyUpdate is the http handler for updating policy data, such as adding users
// TODO: (LowPriority) Write Handler PolicyUpdate
func PolicyUpdate(w http.ResponseWriter, r *http.Request) {

	var p Policy
	var err error
	var requestBody []byte

	// get the user id from the route
	params := httprouter.ParamsFromContext(r.Context())


	requestBody, err = ioutil.ReadAll(r.Body)	

	if err != nil {
		return
	}

	err = json.Unmarshal(requestBody, &p)

	if err != nil {
		return
	}


	p.ID = params.ByName("policyID")
	err = p.update()

	return

}