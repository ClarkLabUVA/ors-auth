package main

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"fmt"
)


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
