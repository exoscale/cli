package egoscale

import (
	"testing"
)

func TestSnapshots(t *testing.T) {
	var _ Taggable = (*Snapshot)(nil)
	var _ AsyncCommand = (*CreateSnapshot)(nil)
	var _ syncCommand = (*ListSnapshots)(nil)
	var _ AsyncCommand = (*DeleteSnapshot)(nil)
	var _ AsyncCommand = (*RevertSnapshot)(nil)
}

func TestSnapshot(t *testing.T) {
	instance := &Snapshot{}
	if instance.ResourceType() != "Snapshot" {
		t.Errorf("ResourceType doesn't match")
	}
}

func TestCreateSnapshot(t *testing.T) {
	req := &CreateSnapshot{}
	if req.name() != "createSnapshot" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*CreateSnapshotResponse)
}

func TestListSnapshots(t *testing.T) {
	req := &ListSnapshots{}
	if req.name() != "listSnapshots" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListSnapshotsResponse)
}

func TestDeleteSnapshot(t *testing.T) {
	req := &DeleteSnapshot{}
	if req.name() != "deleteSnapshot" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanResponse)
}

func TestRevertSnapshot(t *testing.T) {
	req := &RevertSnapshot{}
	if req.name() != "revertSnapshot" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanResponse)
}
