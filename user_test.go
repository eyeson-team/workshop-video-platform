package goose

import (
	"testing"
)

func TestUsers(t *testing.T) {
	setupTestDB()
	defer removeTestDB()

	t.Run("AssignWorkspace", testAssignWorkspace)
	t.Run("AllowedWorkspace", testAllowedWorkspace)
	t.Run("NewUser", testNewUser)
}

func testAssignWorkspace(t *testing.T) {
	user := User{Name: "User Demo"}
	workspace := Workspace{Topic: "Workspace Demo"}
	// user := NewUser("test@eyeson.com")
	db.Create(&user)
	db.Create(&workspace)
	user.Assign(&workspace)

	var reloadedUser User
	db.Preload("Workspaces").First(&reloadedUser, user.ID)
	if len(reloadedUser.Workspaces) != 1 {
		t.Errorf("got %d, expected %d", len(reloadedUser.Workspaces), 1)
	}
}
func testAllowedWorkspace(t *testing.T) {
	user := User{Name: "User Demo"}
	workspace := Workspace{Topic: "Workspace Demo"}
	// user := NewUser("test@eyeson.com")
	db.Create(&user)
	db.Create(&workspace)

	if user.Allowed(&workspace) == true {
		t.Errorf("User has access to non-associated workspace")
	}
	user.Assign(&workspace)

	if user.Allowed(&workspace) == false {
		t.Errorf("User does not have access to associated workspace")
	}
}

func testNewUser(t *testing.T) {
	email := "test@eyeson.com"
	user, err := NewUser(email)
	if err != nil {
		t.Errorf("Failed to create user: %v", err)
	}
	if user.Email != email {
		t.Errorf("got email %v, expected %v", user.Email, email)
	}
	id := user.ID
	user, err = NewUser(email)
	if err != nil {
		t.Errorf("Failed to create user: %v", err)
	}
	if user.ID != id {
		t.Errorf("Got user %v, expected %v", user.ID, id)
	}
}

func TestExtractNameFromEmail(t *testing.T) {
	email := "john.mueller@eyeson.com"
	expected := "John Mueller"
	result := ExtractNameFromEmail(email)

	if result != expected {
		t.Errorf("got %v, expected %v", result, expected)
	}
}
