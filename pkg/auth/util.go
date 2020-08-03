package auth

import (
	"net/http"
	"context"
	"errors"
	"reflect"
	"time"
	mongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

var (
	errNoDocument			= errors.New("NoDocumentFound")
	errDocumentExists       = errors.New("DocumentExists")
	errMongoClient          = errors.New("MongoClientError")
	errMongoQuery           = errors.New("MongoQueryError")
	errMongoDecode          = errors.New("MongoDecodeError")
	errModelFieldValidation = errors.New("ErrorModelFieldValidation")
	errModelMissingField    = errors.New("ErrorModelMissingRequiredField")
	errJSONUnmarshal        = errors.New("ErrorParsingJSON")
	errUUID                 = errors.New("ErrorCreatingUUID")
	errRegex                = errors.New("ErrorRunningRegex")
)

var (
	mongoURI        = os.Getenv("MONGO_URI") 
	mongoDatabase   = os.Getenv("MONGO_DB") 
	mongoCollection = os.Getenv("MONGO_COL")
	orsURI			= os.Getenv("ORS_URI")
)

const (
	typeGroup     = "Organization"
	typeUser      = "Person"
	typeResource  = "Resource"
	typePolicy    = "Policy"
	typeChallenge = "Challenge"
)

func connectMongo() (ctx context.Context, cancel context.CancelFunc, client *mongo.Client, err error) {

	// establish connection with mongo backend
	client, err = mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return
	}

	// create a context for the connection
	ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)

	// connect to the client
	err = client.Connect(ctx)
	return
}

// ErrorDocumentExists determines if an error from the mongo server is a MongoWriteError
// where the document already exists and collides on at least one unique index
// Need to crack open the returned errors
func errorDocumentExists(err error) bool {
	var writeError mongo.WriteError

	// causes pointer error, reflecting on nil causes panic
	if err == nil {
		return false
	}

	// if the mongo operation returned a Write Exception
	errorType := reflect.TypeOf(err)
	errorName := errorType.Name()
	log.Printf("ErrorType: %s, ErrorName: %s", errorType, errorName)

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
