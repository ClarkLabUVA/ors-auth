package main

import (
	"testing"
	"encoding/json"
)

// Basic CRUD Tests Policy
func TestPolicyMethods(t *testing.T) {

	owner := User{Id: "owner"}
	owner.Create()

	u := User{Id: "u1"}
	u.Create()

	r := Resource{Id: "r1", Owner: "owner"}
	r.Create()

	p := Policy{Id: "p1", Resource: "r1"}

	t.Run("Create", func(t *testing.T) {
		err := p.Create()
		if err != nil {
			t.Fatalf("CreatePolicyErr: %s", err.Error())
		}
	})

	t.Run("Get", func(t *testing.T) {
		found := Policy{Id: "p1"}
		err := found.Get()
		if err != nil {
			t.Fatalf("ERROR FindPolicy: %s", err.Error())
		}

		t.Logf("INFO FindPolicy: %+v", found)
	})

	t.Run("List", func(t *testing.T) {
		plist, err := listPolicies()
		if err != nil {
			t.Fatalf("ERROR ListPolicy: %s", err.Error())
		}
		t.Logf("INFO ListPolicy: %+v", plist)
	})

	t.Run("Delete", func(t *testing.T) {
		del := Policy{Id: "p1"}
		err := del.Delete()
		if err != nil {
			t.Fatalf("ERROR DeletePolicy: %s", err.Error())
		}

		t.Logf("INFO DeletePolicy: +%v", del)
	})

	owner.Delete()
	u.Delete()
	r.Delete()


}

func TestPolicyJSONMarshal(t *testing.T) {
	p := Policy{
		Id:        "p1",
		Type:      "Policy",
		Resource:  "r1",
		Principal: []string{"max"},
		Effect:    "Allow",
		Action:    []string{"DeleteIdentifier"},
		Issuer:    "ors:mds",
	}

	pJSON, err := json.Marshal(p)

	if err != nil {
		t.Fatalf("ERROR: %s", err.Error())
	}

	t.Logf("MarshaledGroup: %s", string(pJSON))


}

func TestPolicyJSONUnmarshal(t *testing.T) {}
