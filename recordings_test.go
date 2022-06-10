package goose

import "testing"

func TestRecordnigsStoragePath(t *testing.T) {
	rec := Recording{ID: 3, WorkspaceID: 8}
	expectedPath := "/tmp/meetduck/8-3.webm"
	if rec.StoragePath() != expectedPath {
		t.Errorf("StoragePath() = %s, want %s", rec.StoragePath(), expectedPath)
	}
}
