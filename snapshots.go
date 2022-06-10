package goose

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Snapshot struct {
	ID          uint
	Reference   string
	CreatedAt   time.Time
	WorkspaceID uint
	Workspace   Workspace
}

func (s *Snapshot) Path() string {
	return fmt.Sprintf("/snapshots/%d.jpg", s.ID)
}

// StoragePath determines the local storage path for the snapshot.
func (s *Snapshot) StoragePath() string {
	varDir := VARDIR
	if _, err := os.Stat(varDir); os.IsNotExist(err) {
		os.MkdirAll(varDir, 0755)
	}

	return fmt.Sprintf("%s/%d-%d.jpg", varDir, s.WorkspaceID, s.ID)
}

// Store downloads file and stores it in our storage path.
func (s *Snapshot) Store(downloadLink string) error {
	dest, err := os.Create(s.StoragePath())
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
