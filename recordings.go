package goose

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Recording struct {
	ID          uint
	Reference   string
	Duration    uint
	StartedAt   time.Time
	EndedAt     time.Time
	WorkspaceID uint
	Workspace   Workspace
}

func (r *Recording) Path() string {
	return fmt.Sprintf("/recordings/%d.webm", r.ID)
}

// StoragePath determines the local storage path for the recording.
func (r *Recording) StoragePath() string {
	varDir := VARDIR
	if _, err := os.Stat(varDir); os.IsNotExist(err) {
		os.MkdirAll(varDir, 0755)
	}

	return fmt.Sprintf("%s/%d-%d.webm", varDir, r.WorkspaceID, r.ID)
}

// Store downloads file and stores it in our storage path.
func (r *Recording) Store(downloadLink string) error {
	dest, err := os.Create(r.StoragePath())
	if err != nil {
		return err
	}
	defer dest.Close()

	resp, err := http.Get(downloadLink)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to download. Status %d", resp.StatusCode)
	}

	_, err = io.Copy(dest, resp.Body)
	return err
}
