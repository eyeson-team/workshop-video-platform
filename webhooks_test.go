package goose

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	eyeson "github.com/eyeson-team/eyeson-go"
)

func readWebhookSample(filename string) (*eyeson.Webhook, error) {
	sample, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var webhook eyeson.Webhook
	if err = json.Unmarshal(sample, &webhook); err != nil {
		return nil, err
	}
	return &webhook, nil
}

func TestWebhooks(t *testing.T) {
	setupTestDB()
	defer removeTestDB()

	t.Run("CreateMeeting", testWebhookCreatesMeeting)
	t.Run("StoreMeetingEnd", testWebhookStoreMeetingEnd)
	t.Run("CreateRecording", testWebhookCreatesRecording)
	t.Run("StoreRecordingReference", testWebhookStoreRecordingReference)
	t.Run("CreateSnapshot", testWebhookCreatesSnapshot)
}

func testWebhookCreatesMeeting(t *testing.T) {
	webhook, err := readWebhookSample("./fixtures/webhook_room_update.json")
	if err != nil {
		t.Errorf("Failed to read webhook fixture: %v", err)
	}
	workspace := Workspace{Topic: "webhook creates meeting"}
	if err := db.Create(&workspace).Error; err != nil {
		t.Errorf("Failed to create workspace, %v", err)
	}
	webhook.Room.Id = fmt.Sprintf("%d", workspace.ID)

	var beforeCount, afterCount int64
	db.Model(&Meeting{}).Count(&beforeCount)
	if err := storeWebhookEvent(webhook); err != nil {
		t.Errorf("Failed to store webhook, %v", err)
	}
	db.Model(&Meeting{}).Count(&afterCount)
	if afterCount == beforeCount {
		t.Errorf("got %d, expected %d", afterCount, beforeCount+1)
	}
}

func testWebhookStoreMeetingEnd(t *testing.T) {
	webhook, err := readWebhookSample("./fixtures/meeting_ended.json")
	if err != nil {
		t.Errorf("Failed to read webhook fixture: %v", err)
	}
	workspace := Workspace{Topic: "webhook store meeting ended"}
	if err := db.Create(&workspace).Error; err != nil {
		t.Errorf("Failed to create workspace, %v", err)
	}
	webhook.Room.Id = fmt.Sprintf("%d", workspace.ID)
	meeting := Meeting{WorkspaceID: workspace.ID, StartedAt: time.Now()}
	if err := db.Create(&meeting).Error; err != nil {
		t.Errorf("Failed to create meeting, %v", err)
	}

	storeWebhookEvent(webhook)
	db.Last(&meeting)
	if meeting.Ended() != true {
		t.Errorf("got %v, expected %v", meeting.Ended(), true)
	}
}

func testWebhookCreatesRecording(t *testing.T) {
	webhook, err := readWebhookSample("./fixtures/webhook_recording_update.json")
	if err != nil {
		t.Errorf("Failed to read webhook fixture: %v", err)
	}
	var workspace Workspace
	if err := db.FirstOrCreate(&workspace).Error; err != nil {
		t.Errorf("Failed to create workspace, %v", err)
	}
	webhook.Recording.Room.Id = fmt.Sprintf("%d", workspace.ID)

	var beforeCount, afterCount int64
	db.Model(&Recording{}).Count(&beforeCount)
	storeWebhookEvent(webhook)
	db.Model(&Recording{}).Count(&afterCount)
	if afterCount == beforeCount {
		t.Errorf("got %d, expected %d", afterCount, afterCount+1)
	}
}

func testWebhookStoreRecordingReference(t *testing.T) {
	webhook, err := readWebhookSample("./fixtures/webhook_recording_update.json")
	if err != nil {
		t.Errorf("Failed to read webhook fixture: %v", err)
	}
	storeWebhookEvent(webhook)

	var recording Recording
	result := db.Model(&Recording{}).Last(&recording)
	if result.Error != nil {
		t.Errorf("Failed to fetch last recording, got %v", result.Error)
	}
	if recording.Reference != webhook.Recording.Id {
		t.Errorf("got %v, expected %v", recording.Reference, webhook.Recording.Id)
	}
}

func testWebhookCreatesSnapshot(t *testing.T) {
	webhook, err := readWebhookSample("./fixtures/webhook_snapshot_update.json")
	if err != nil {
		t.Errorf("Failed to read webhook fixture: %v", err)
	}
	var workspace Workspace
	if err := db.FirstOrCreate(&workspace).Error; err != nil {
		t.Errorf("Failed to create workspace, %v", err)
	}
	webhook.Snapshot.Room.Id = fmt.Sprintf("%d", workspace.ID)

	var beforeCount, afterCount int64
	db.Model(&Snapshot{}).Count(&beforeCount)
	storeWebhookEvent(webhook)
	db.Model(&Snapshot{}).Count(&afterCount)
	if afterCount == beforeCount {
		t.Errorf("got %d, expected %d", afterCount, afterCount+1)
	}
}
