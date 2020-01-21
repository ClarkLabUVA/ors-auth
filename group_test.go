package main

import (
	"testing"
	"encoding/json"
)

func init() {
	// drop collection
	MongoCollection = "test"

	mongoCtx, cancel, client, _ := connectMongo()
	defer cancel()

	client.Database(MongoDatabase).Collection(MongoCollection).Drop(mongoCtx)
}

// Basic CRUD Tests for Groups
func TestGroupMethods(t *testing.T) {
	var err error

	admin := User{
		Id: "orcid:1",
		Email: "admin@gmail.com",
		Type: TypeUser,
		Groups: []string{},
	}

	err = admin.Create()

	if err != nil {
		t.Fatalf("SetupFailure: Failed to Create Admin\t Error: %s", err.Error())
	}

	member := User{
		Id: "member",
		Email: "member@gmail.com",
		Type: TypeUser,
		Groups: []string{},
	}
	err = member.Create()

	if err != nil {
		t.Fatalf("SetupFailure: Failed to Create User\t Error: %s", err.Error())
	}

	g := Group{
		Id: "group1",
		Type: TypeGroup,
		Admin: "orcid:1",
		Members: []string{"orcid:2"},
	}

	t.Run("Create", func(t *testing.T) {

		err := g.Create()
		if err != nil {
			t.Fatalf("Failed to Create Group: %s", err.Error())
		}

		// Verify Admin has member of group
		var updatedAdmin User
		updatedAdmin.Id = admin.Id
		err = updatedAdmin.Get()

		if err != nil {
			t.Fatalf("Failed to Fetch Updated Admin: %s", err.Error())
		}

		if len(updatedAdmin.Groups) != 1 {
			t.Fatalf("Admin is not listed as member of group: %+v", updatedAdmin)
		}

		// Verify User is member of group
		var updatedUser User
		updatedUser.Id = member.Id
		err = updatedUser.Get()

		if err != nil {
			t.Fatalf("Failed to Fetch Updated User: %s", err.Error())
		}

	})

	t.Run("Get", func(t *testing.T) {
		found := Group{Id: "group1"}
		err := found.Get()
		if err != nil {
			t.Fatalf("Failed to Find Group: %s", err.Error())
		}
	})

	t.Run("List", func(t *testing.T) {
		_, err := listGroups()
		if err != nil {
			t.Fatalf("Failed to List Groups: %s", err.Error())
		}
	})

	t.Run("Delete", func(t *testing.T) {
		del := Group{Id: "group1"}
		err := del.Delete()
		if err != nil {
			t.Fatalf("Failed to Delete Group: %s", err.Error())
		}

		t.Logf("Deleted Group: %+v", del)
	})


	admin.Delete()
	member.Delete()

}


func TestGroupJSONUnmarshal(t *testing.T) {

	t.Run("Valid", func(t *testing.T) {
		//var g Group
		//groupBytes := []byte(`{}`)
		// assure admin is listed as member
	})
	t.Run("MissingName", func(t *testing.T) {})

}

func TestGroupJSONMarshal(t *testing.T) {

	g := Group{
		Id:      "test_group",
		Type:    "Group",
		Name:    "test_group",
		Admin:   "max",
		Members: []string{"u1", "u2"},
	}

	groupJSON, err := json.Marshal(g)

	if err != nil {
		t.Fatalf("ERROR: %s", err.Error())
	}

	t.Logf("MarshaledGroup: %s", string(groupJSON))

}
