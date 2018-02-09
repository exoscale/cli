package egoscale

import (
	"testing"
)

func TestSnapshots(t *testing.T) {
	var _ Taggable = (*Snapshot)(nil)
	var _ asyncCommand = (*CreateSnapshot)(nil)
	var _ syncCommand = (*ListSnapshots)(nil)
	var _ asyncCommand = (*DeleteSnapshot)(nil)
	var _ asyncCommand = (*RevertSnapshot)(nil)
}

func TestSnapshot(t *testing.T) {
	instance := &Snapshot{}
	if instance.ResourceType() != "Snapshot" {
		t.Errorf("ResourceType doesn't match")
	}
}

func TestCreateSnapshot(t *testing.T) {
	req := &CreateSnapshot{}
	if req.APIName() != "createSnapshot" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*CreateSnapshotResponse)
}

func TestListSnapshots(t *testing.T) {
	req := &ListSnapshots{}
	if req.APIName() != "listSnapshots" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListSnapshotsResponse)
}

func TestDeleteSnapshot(t *testing.T) {
	req := &DeleteSnapshot{}
	if req.APIName() != "deleteSnapshot" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestRevertSnapshot(t *testing.T) {
	req := &RevertSnapshot{}
	if req.APIName() != "revertSnapshot" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}
