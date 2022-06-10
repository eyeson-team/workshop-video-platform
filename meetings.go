package goose

import (
	"time"
)

// Meeting does represent a online video conference and is created from the
// webhook information received by the eyeson API service.
type Meeting struct {
	ID          uint
	Content     string
	StartedAt   time.Time
	EndedAt     time.Time
	WorkspaceID uint
	Workspace   Workspace
}

// Active provides information if a meeting is still active.
func (m *Meeting) Active() bool {
	return m.EndedAt.IsZero()
}

// Ended provides information if the meeting already has been ended.
func (m *Meeting) Ended() bool {
	return m.Active() == false
}

// StoreShutdown sets the ended_at timestamp to the current time.
func (m *Meeting) StoreShutdown() error {
	m.EndedAt = time.Now().UTC()
	return db.Save(&m).Error
}

// NewMeeting does create a meeting if it does not exist.
// TODO: should also return the newly created meeting
func NewMeeting(workspace *Workspace, startedAt time.Time) error {
	meeting := Meeting{Workspace: *workspace, StartedAt: startedAt}
	qry := db.Where("started_at = ? AND workspace_id = ?", startedAt, workspace.ID)
	return qry.FirstOrCreate(&meeting).Error
}
