package goose

type Workspace struct {
	ID         uint
	Topic      string
	Content    string
	Meetings   []Meeting
	Recordings []Recording
	Snapshots  []Snapshot
	Users      []*User `gorm:"many2many:workspace_users;"`
}

// LastMeeting provides the last meeting of a workspace.
func (w *Workspace) LastMeeting() (*Meeting, error) {
	var meeting Meeting
	result := db.Where("workspace_id = ?", w.ID).Last(&meeting)
	return &meeting, result.Error
}
