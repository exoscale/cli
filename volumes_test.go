package egoscale

import (
	"testing"
)

func TestVolumes(t *testing.T) {
	var _ Taggable = (*Volume)(nil)
	var _ Command = (*ListVolumes)(nil)
	var _ AsyncCommand = (*ResizeVolume)(nil)
}

func TestVolume(t *testing.T) {
	instance := &Volume{}
	if instance.ResourceType() != "Volume" {
		t.Errorf("ResourceType doesn't match")
	}
}

func TestListVolumes(t *testing.T) {
	req := &ListVolumes{}
	if req.APIName() != "listVolumes" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListVolumesResponse)
}

func TestResizeVolume(t *testing.T) {
	req := &ResizeVolume{}
	if req.APIName() != "resizeVolume" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*ResizeVolumeResponse)
}
