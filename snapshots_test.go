package goose

import "testing"

func TestSnapshotStoragePath(t *testing.T) {
	snap := Snapshot{ID: 7, WorkspaceID: 42}
	expectedPath := "/tmp/meetduck/42-7.jpg"
	if snap.StoragePath() != expectedPath {
		t.Errorf("StoragePath() = %s, want %s", snap.StoragePath(), expectedPath)
	}
}
