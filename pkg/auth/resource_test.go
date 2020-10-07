package auth

import (
	"testing"
	"encoding/json"
)

func TestResourceMethods(t *testing.T) {

	u := User{Id: "owner"}

	r := Resource{Id: "res1", Owner: u.Id}

	t.Run("Create", func(t *testing.T) {
		err := r.Create()
		if err != nil {
			t.Fatalf("Failed to Create Resource: %s", err.Error())
		}
	})

	t.Run("Get", func(t *testing.T) {
		found := Resource{Id: "res1"}
		err := found.Get()
		if err != nil {
			t.Fatalf("Failed to Find Resource: %s", err.Error())
		}
	})

	t.Run("List", func(t *testing.T) {
		rlist, err := listResources()
		if err != nil {
			t.Fatalf("Failed to List Resources: %s", err.Error())
		}
		t.Logf("ListResources: %+v", rlist)
	})

	t.Run("Delete", func(t *testing.T) {
		del := Resource{Id: "res1"}
		err := del.Delete()
		if err != nil {
			t.Fatalf("DeleteResourceError: %s", err.Error())
		}
		t.Logf("DeleteResource: %+v", del)
	})

}


func TestResourceJSONUnmarshal(t *testing.T) {}

func TestResourceJSONMarshal(t *testing.T) {

	r := Resource{
		Id:    "resource1",
		Type:  "resource",
		Owner: "max",
	}

	resJSON, err := json.Marshal(r)

	if err != nil {
		t.Fatalf("ERROR: %s", err.Error())
	}

	t.Logf("MarshaledResource: %s", string(resJSON))

}
