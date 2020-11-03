package auth

import (
	"testing"
	"encoding/json"
)

func init() {
	// drop collection
	mongoDatabase = "test"
	mongoCollection = "test"
	mongoURI = "mongodb://mongouser:mongosecret@127.0.0.1:27017"

	mongoCtx, cancel, client, _ := connectMongo()
	defer cancel()

	client.Database(mongoDatabase).Collection(mongoCollection).Drop(mongoCtx)
}

// Basic CRUD Tests for Groups
func TestGroupMethods(t *testing.T) {
	var err error

	admin := User{
		ID: "orcid:1",
		Email: "admin@gmail.com",
		Type: typeUser,
		Groups: []string{},
	}

	err = admin.create()

	if err != nil {
		t.Fatalf("SetupFailure: Failed to Create Admin\t Error: %s", err.Error())
	}

	member := User{
		ID: "member",
		Email: "member@gmail.com",
		Type: typeUser,
		Groups: []string{},
	}
	err = member.create()

	if err != nil {
		t.Fatalf("SetupFailure: Failed to Create User\t Error: %s", err.Error())
	}

	g := Group{
		ID: "group1",
		Type: typeGroup,
		Admin: admin.ID,
		Members: []string{member.ID},
	}

	t.Run("Create", func(t *testing.T) {

		err := g.create()
		if err != nil {
			t.Fatalf("Failed to Create Group: %s", err.Error())
		}

		// Verify Admin has member of group
		var updatedAdmin User
		updatedAdmin.ID = admin.ID
		err = updatedAdmin.get()

		if err != nil {
			t.Fatalf("Failed to Fetch Updated Admin: %s", err.Error())
		}

		if len(updatedAdmin.Groups) != 1 {
			t.Fatalf("Admin is not listed as member of group: %+v", updatedAdmin)
		}

		// Verify User is member of group
		var updatedUser User
		updatedUser.ID = member.ID
		err = updatedUser.get()

		if err != nil {
			t.Fatalf("Failed to Fetch Updated User: %s", err.Error())
		}

	})

	t.Run("Get", func(t *testing.T) {
		found := Group{ID: "group1"}
		err := found.get()
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
		del := Group{ID: "group1"}
		err := del.delete()
		if err != nil {
			t.Fatalf("Failed to Delete Group: %s", err.Error())
		}

		t.Logf("Deleted Group: %+v", del)
	})


	// TODO does the user session obtain the new groups


	// Clean up test
	/*
	admin.delete()
	member.delete()
	*/


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
		ID:      "test_group",
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
