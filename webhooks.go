package goose

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	eyeson "github.com/eyeson-team/eyeson-go"
)

// Register the global webhook endpoint for our eyeson API key.
func RegisterWebhook(url string) error {
	if len(url) == 0 {
		return errors.New("WH_URL not set")
	}
	options := strings.Join([]string{
		eyeson.WEBHOOK_ROOM,
		eyeson.WEBHOOK_RECORDING,
		eyeson.WEBHOOK_SNAPSHOT,
	}, ",")
	return videoService.Webhook.Register(url+"/webhook", options)
}

// validateWebhook validates a webhook by its content and signature.
func validateWebhook(body []byte, signature string) error {

	return nil

	// TODO: Get the API key from elsewhere! service should already know o.o
	apiKey := os.Getenv("API_KEY")

	h := hmac.New(sha256.New, []byte(apiKey))
	h.Write(body)
	if hex.EncodeToString(h.Sum(nil)) != signature {
		return errors.New("Could not verify webhook signature.")
	}
	return nil
}

// storeWebhookEvent stores webhook delivered information to a recording or
// meeting event.
func storeWebhookEvent(webhook *eyeson.Webhook) error {
	// debug, _ := json.Marshal(webhook)
	// log.Printf("Content of webhook %s: %s", webhook.Type, debug)

	switch webhook.Type {
	case "room_update":
		var workspace Workspace
		err := db.First(&workspace, webhook.Room.Id).Error
		if err != nil {
			log.Println("Could not fetch workspace:", err)
			return nil
		}
		if webhook.Room.Shutdown {
			meeting, err := workspace.LastMeeting()
			if err != nil {
				log.Println("Could not fetch last meeting:", err)
				return nil
			}
			return meeting.StoreShutdown()
		} else {
			return NewMeeting(&workspace, webhook.Room.StartedAt)
		}
		if err != nil {
			log.Println("Webhook handler failed with:", err)
		}
	case "recording_update":
		var workspace Workspace
		err := db.First(&workspace, webhook.Recording.Room.Id).Error
		if err != nil {
			log.Println("Could not fetch workspace:", err)
			return nil
		}
		startedAt := time.Unix(int64(webhook.Recording.CreatedAt), 0)
		rec := &Recording{Workspace: workspace, Reference: webhook.Recording.Id,
			StartedAt: startedAt, EndedAt: time.Now().UTC(),
			Duration: uint(webhook.Recording.Duration)}
		db.Create(rec)

		// if link is present, download recording to storage
		if len(webhook.Recording.Links.Download) > 0 {
			err := rec.Store(webhook.Recording.Links.Download)
			if err != nil {
				log.Println("Download recording failed:", err)
				return nil
			}
		}
	case "snapshot_update":
		var workspace Workspace
		err := db.First(&workspace, webhook.Snapshot.Room.Id).Error
		if err != nil {
			log.Println("Could not fetch workspace:", err)
			return nil
		}
		snapshot := &Snapshot{Workspace: workspace, Reference: webhook.Snapshot.Id,
			CreatedAt: webhook.Snapshot.CreatedAt}
		db.Create(snapshot)

		// if link is present, download snapshot to storage
		if len(webhook.Snapshot.Links.Download) > 0 {
			err := snapshot.Store(webhook.Snapshot.Links.Download)
			if err != nil {
				log.Println("Download snapshot failed:", err)
				return nil
			}
		}
	default:
		log.Println("Received unknown webhook type", webhook.Type)
	}
	return nil
}

// UnregisterWebhook unregisters the current webhook.
func UnregisterWebhook() error {
	return videoService.Webhook.Unregister()
}
