package auth


import (
	"testing"
	"time"
	"encoding/json"
)

// Basic CRUD Tests for Challneges
func TestChallengeMethods(t *testing.T) {

	owner := User{Id: "owner"}
	owner.Create()

	u := User{Id: "u1"}
	u.Create()

	r := Resource{Id: "r1", Owner: "owner"}
	r.Create()

	p := Policy{
		Id:        "p1",
		Resource:  "r1",
		Principal: []string{"u1"},
		Action:    []string{"read"},
		Effect:    "Allow",
	}
	p.Create()

	t.Run("Evaluate", func(t *testing.T) {

		t.Run("Owner", func(t *testing.T) {

			c := Challenge{
				Principal: "owner",
				Resource:  "r1",
				Action:    "delete",
			}

			c.Evaluate()

			if !c.Granted {
				t.Fatalf("ERROR ChallengeOwner: Owner of Resource Wrongly Denied Permission")
			}
		})

		t.Run("PolicyAllowed", func(t *testing.T) {
			c := Challenge{
				Principal: "u1",
				Resource:  "r1",
				Action:    "read",
			}

			err := c.Evaluate()
			if err != nil {
				t.Fatalf("ERROR ChallengeEvaluation: %s", err.Error())
			}

			if !c.Granted {
				t.Fatalf("ERROR Challenge Incorrectly Denied")
			}

		})

		t.Run("PolicyMissingAction", func(t *testing.T) {

			c := Challenge{
				Principal: "u1",
				Resource:  "r1",
				Action:    "write",
			}

			err := c.Evaluate()
			if err != nil {
				t.Fatalf("ERROR ChallengeEvaluation: %s", err.Error())
			}

			if c.Granted {
				t.Fatalf("ERROR Challenge Incorrectly Granted")
			}

		})

	})

	t.Run("List", func(t *testing.T) {
		clist, err := listChallenges()
		if err != nil {
			t.Fatalf("ERROR ListChallenges: %s", err.Error())
		}
		t.Logf("INFO ListChallenges: %+v", clist)

	})
}

func TestChallengeJSONUnmarshal(t *testing.T) {}

func TestChallengeJSONMarshal(t *testing.T) {
	c := Challenge{
		Id:        "c1",
		Type:      "Challenge",
		Principal: "max",
		Resource:  "r1",
		Action:    "DeleteIdentifer",
		Time:      time.Now(),
		Issuer:    "ors:mds",
		Granted:   true,
	}

	chalJSON, err := json.Marshal(c)

	if err != nil {
		t.Fatalf("ERROR: %s", err.Error())
	}

	t.Logf("MarshaledGroup: %s", string(chalJSON))

}
