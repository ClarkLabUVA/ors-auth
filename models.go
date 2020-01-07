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
