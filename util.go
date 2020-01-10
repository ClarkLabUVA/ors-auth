package main

import (
	"net/http"
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"
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

func MongoFindOne(Id string) (b []byte, err error) {
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

func MongoDeleteOne(Id string) (b []byte, err error) {
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

// Function to determine if an error from the mongo server is a MongoWriteError
// where the document already exists and collides on at least one unique index
// Need to crack open the returned errors
func ErrorDocumentExists(err error) bool {
	var writeError mongo.WriteError

	// if the mongo operation returned a Write Exception
	errorName := errorType.Name()

	if errorName == "WriteErrors" {
		writeError = err.(mongo.WriteErrors)[0]
	}
	if errorName == "WriteException" {
		writeError = err.(mongo.WriteException).WriteErrors[0]
	}

	if errorName != "WriteErrors" && errorName != "WriteException" {
		return false
	}

	if writeError.Code == 11000 {
		return true
	}

	return false
}

// TODO: (LowPriority) Write a Generic Error Handler to return a JSON error
func handleErrors(err error, w http.ResponseWriter, r *http.Request) {
	//

}
