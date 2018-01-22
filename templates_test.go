package egoscale

import (
	"testing"
)

func TestTemplates(t *testing.T) {
	var _ Taggable = (*Template)(nil)
	var _ Command = (*ListTemplates)(nil)
}

func TestTemplate(t *testing.T) {
	instance := &Template{}
	if instance.ResourceType() != "Template" {
		t.Errorf("ResourceType doesn't match")
	}
}

func TestListTemplates(t *testing.T) {
	req := &ListTemplates{}
	if req.name() != "listTemplates" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListTemplatesResponse)
}
