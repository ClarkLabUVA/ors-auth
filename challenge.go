package main

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"fmt"
)

type Challenge struct {
	Id        string    `json:"@id" bson:"@id"`
	Type      string    `json:"@type" bson:"@type"`
	Principal string    `json:"principal" bson:"principal"`
	Resource  string    `json:"resource" bson:"resource"`
	Action    string    `json:"action" bson:"action"`
	Time      time.Time `json:"time" bson:"time"`
	Issuer    string    `json:"issuer" bson:"issuer"`
	Granted   bool      `json:"granted" bson:"granted"`
}

func (c Challenge) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	var err error

	// open quotes
	buf.WriteString(`{`)

	// write context
	buf.WriteString(`"@context": {"@base": "http://schema.org/", "principal": "http://schema.org/agent"}, "@type": "AuthorizeAction", `)

	// write out challenge id as full url
	buf.WriteString(fmt.Sprintf(`"@id": "%schallenge/%s", `, ORSURI, c.Id))

	// write principal
	buf.WriteString(fmt.Sprintf(`"principal": "%s", `, c.Principal))

	// write resource
	buf.WriteString(fmt.Sprintf(`"resource": "%s", `, c.Resource))

	// write action
	buf.WriteString(fmt.Sprintf(`"action": "%s", `, c.Action))

	// write time
	jsonTime, _ := c.Time.MarshalJSON()
	buf.WriteString(fmt.Sprintf(`"time": %s, `, string(jsonTime)))

	// write granted
	buf.WriteString(fmt.Sprintf(`"granted": %t }`, c.Granted))

	out := buf.Bytes()
	return out, err
}

func (c *Challenge) UnmarshalJSON(data []byte) error {
	var err error

	type Alias Challenge
	aux := struct {
		Type    string    `json:"@type" bson:"@type"`
		Time    time.Time `json:"time" bson:"time"`
		Granted bool      `json:"granted" bson:"granted"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	err = json.Unmarshal(data, &aux)

	if err != nil {
		return err
	}

	challengeId, err := uuid.NewUUID()

	if err != nil {
		return err
	}

	c.Id = challengeId.String()

	c.Type = TypeChallenge
	c.Time = time.Now()
	c.Granted = false

	return nil
}

func (c Challenge) ID() string {
	return c.Id
}

func (c *Challenge) Evaluate() (err error) {

	c.Type = "Challenge"
	c.Time = time.Now()
	c.Granted = false

	// connect to the client
	ctx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		return fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
	}

	// connect to collection
	collection := client.Database(MongoDatabase).Collection(MongoCollection)

	// get principal
	var u User
	err = collection.FindOne(ctx, bson.D{{"@id", c.Principal}}).Decode(&u)

	// get resource
	var r Resource
	err = collection.FindOne(ctx, bson.D{{"@id", c.Resource}}).Decode(&r)

	// if principal is owner
	if c.Principal != r.Owner {

		// search policies for granting relevant permissions
		var matchingPolicies []Policy

		// change to a regex match for the action
		// i.e. action s3:download can be matched by s3:*
		var cursor *mongo.Cursor
		cursor, err = collection.Find(ctx,
			bson.D{
				{"@type", "Policy"},
				{"effect", "Allow"},
				{"resource", c.Resource},
				{"action", bson.D{{"$elemMatch",
					bson.D{{"$in", []string{"*", c.Action}}}}}},
				{"principal", bson.D{{"$elemMatch",
					bson.D{{"$in", append(u.Groups, u.Id)}}}}},
			})

		if err != nil {
			return
		}

		err = cursor.All(ctx, &matchingPolicies)

		if err != nil {
			return
		}

		if len(matchingPolicies) == 0 {
			c.Granted = false
		} else {
			c.Granted = true
		}

	} else {
		c.Granted = true
	}

	_, err = collection.InsertOne(ctx, c)
	return

}

func listChallenges() (c []Challenge, err error) {

	mongoCtx, cancel, client, err := connectMongo()
	defer cancel()

	if err != nil {
		err = fmt.Errorf("%w: %s", ErrMongoClient, err.Error())
		return
	}

	collection := client.Database(MongoDatabase).Collection(MongoCollection)

	query := bson.D{{"@type", "Challenge"}}
	cur, err := collection.Find(mongoCtx, query, nil)
	defer cur.Close(mongoCtx)
	if err != nil {
		return
	}

	err = cur.All(mongoCtx, &c)
	if err != nil {
		return
	}

	return

}


func ChallengeEvaluate(w http.ResponseWriter, r *http.Request) {

	var c Challenge
	var err error
	var requestBody []byte

	requestBody, err = ioutil.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error": "%w"}`, err)
		return
	}

	if err = json.Unmarshal(requestBody, &c); err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error": "%w"}`, err)
		return
	}

	err = c.Evaluate()
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error": "%w"}`, err)
		return
	}

	if c.Granted {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(403)
	}

	fmt.Fprintf(w, `{"@id": "%s","granted": "%t"}`, c.Id, c.Granted)
}

func ChallengeList(w http.ResponseWriter, r *http.Request) {

	var challengeList []Challenge
	var responseBody []byte
	var err error

	w.Header().Set("Content-Type", "application/ld+json")

	challengeList, err = listChallenges()

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, `{"error": "%w"}`, err)
		return
	}

	responseBody, err = json.Marshal(challengeList)

	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error": "%w"}`, err)
		return
	}

	w.WriteHeader(200)
	w.Write(responseBody)

}
