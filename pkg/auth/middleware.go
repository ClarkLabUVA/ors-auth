package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	log "github.com/sirupsen/logrus"
	"context"
	"time"
	"github.com/google/uuid"
	"os"
)

var (
	AdminPassword string
	AdminUser	string
)

func init() {

  // Log as JSON instead of the default ASCII formatter.
  log.SetFormatter(&log.JSONFormatter{})

  // Output to stdout instead of the default stderr
  // Can be any io.Writer, see below for File example
  log.SetOutput(os.Stdout)

  // Only log the warning severity or above.
  log.SetLevel(log.InfoLevel)

}

func BasicAuth(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Get the Basic Authentication credentials
	user, password, hasAuth := r.BasicAuth()

	if hasAuth && user == AdminUser && password == AdminPassword {
		// Delegate request to the given handle
		next(rw, r)
	}

	// Request Basic Authentication otherwise
	rw.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
	http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

}

func ValidJSON(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	var err error
	var requestBody []byte

	requestBody, err = ioutil.ReadAll(r.Body)

	if err != nil {
		rw.WriteHeader(400)
		rw.Header().Set("Content-Type", "application/ld+json")
		rw.Write([]byte(`{"error": "Unable to Read Request Body"}`))
		return
	}

	// If Error for Unmarshaling JSON Body
	if !json.Valid(requestBody) {
		rw.WriteHeader(400)
		rw.Header().Set("Content-Type", "application/ld+json")
		rw.Write([]byte(`{"error": "Invalid JSON Submitted"}`))
		return
	}

	next(rw, r)
}

// Add a logrus logger with default messages into the request context
func LoggingMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	requestId := uuid.New()

	contextLogger := log.WithFields(log.Fields{
    "@id": requestId.String(),
    "time": time.Now(),
  })

	ctx := context.WithValue(r.Context(), "log", contextLogger)

	next(rw, r.WithContext(ctx))
}
