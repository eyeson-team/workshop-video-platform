package goose

import "testing"

func TestWorkspace(t *testing.T) {
	setupTestDB()
	defer removeTestDB()

	t.Run("CreateMeeting", testWorkspaceLastMeeting)
}

func testWorkspaceLastMeeting(t *testing.T) {
	w := Workspace{Topic: "test"}
	db.Create(&w)
	m := Meeting{WorkspaceID: w.ID}
	db.Create(&m)
	last, err := w.LastMeeting()
	if err != nil {
		t.Errorf("Could not fetch last meeting, %v.", err)
	}
	if last.ID != m.ID {
		t.Errorf("Got meeting id %d, wanted %d", last.ID, m.ID)
	}
}
