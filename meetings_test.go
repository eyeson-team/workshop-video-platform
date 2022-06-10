package goose

import (
	"testing"
	"time"
)

func TestMeetings(t *testing.T) {
	setupTestDB()
	defer removeTestDB()

	t.Run("Create", testCreateMeeting)
	t.Run("Update", testUpdateMeeting)
	t.Run("End", testEndMeeting)
	t.Run("Uniqueness", testMeetingUniqueness)
}

func testCreateMeeting(t *testing.T) {
	workspace := Workspace{Topic: "test meeting create"}
	db.Create(&workspace)

	if err := NewMeeting(&workspace, time.Now()); err != nil {
		t.Errorf("Meeting create failed with %v", err)
	}

	var meeting Meeting
	if err := db.Last(&meeting).Error; err != nil {
		t.Errorf("Created meeting was not found with error %v", err)
	}
}

func testUpdateMeeting(t *testing.T) {
	workspace := Workspace{Topic: "test meeting update"}
	db.Create(&workspace)
	meeting := Meeting{Workspace: workspace, StartedAt: time.Now()}
	db.Create(&meeting)

	if err := NewMeeting(&workspace, meeting.StartedAt); err != nil {
		t.Errorf("Meeting update failed with %v", err)
	}

	if meeting.EndedAt.After(meeting.StartedAt) {
		t.Errorf("got %v expected to be after %v", meeting.EndedAt, meeting.StartedAt)
	}
}

func testEndMeeting(t *testing.T) {
	workspace := Workspace{Topic: "test meeting update"}
	db.Create(&workspace)
	meeting := Meeting{Workspace: workspace, StartedAt: time.Now()}
	db.Create(&meeting)

	if err := meeting.StoreShutdown(); err != nil {
		t.Errorf("Meeting update failed with %v", err)
	}
	var result Meeting
	if err := db.Where("workspace_id = ?", workspace.ID).Last(&result).Error; err != nil {
		t.Errorf("Meeting fetch failed with %v", err)
	}
	if result.Ended() != true {
		t.Errorf("got %v, expected %v", result.Ended(), true)
	}
}

func testMeetingUniqueness(t *testing.T) {
	workspace := Workspace{Topic: "test meeting uniqueness"}
	if err := db.Create(&workspace).Error; err != nil {
		t.Errorf("Failed to create workspace with %v", err)
	}

	startTime := time.Now()
	if err := NewMeeting(&workspace, startTime); err != nil {
		t.Errorf("Meeting create failed with %v", err)
	}
	if err := NewMeeting(&workspace, startTime); err != nil {
		t.Errorf("Meeting create failed with %v", err)
	}
	db.Preload("Meetings").First(&workspace, workspace.ID) // reload
	if len(workspace.Meetings) != 1 {
		t.Errorf("got %d, want %d", len(workspace.Meetings), 1)
	}
}
