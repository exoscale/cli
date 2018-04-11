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
	ts := newServer(response{200, `
		
		{ "listtemplateresponse": {
			"count": 1,
			"template": [
			  {
				"account": "exostack",
				"checksum": "3c80c71fcfe1e2e88c12ca7d39c294a0",
				"created": "2018-01-30T09:16:05+0100",
				"crossZones": false,
				"details": {
				  "username": "debian"
				},
				"displaytext": "Linux Debian 9 64-bit 200G Disk (2018-01-18-25e9ec)",
				"domain": "ROOT",
				"domainid": "4a8857b8-7235-4e31-a7ef-b8b44d180850",
				"format": "QCOW2",
				"hypervisor": "KVM",
				"id": "a8a4b773-32ce-4d1c-a19b-21da055ec5c6",
				"isdynamicallyscalable": false,
				"isextractable": false,
				"isfeatured": true,
				"ispublic": true,
				"isready": true,
				"name": "Linux Debian 9 64-bit",
				"ostypeid": "113038d0-a8cd-4d20-92be-ea313f87c3ac",
				"ostypename": "Other PV (64-bit)",
				"passwordenabled": true,
				"size": 214748364800,
				"sshkeyenabled": false,
				"tags": [],
				"templatetype": "USER",
				"zoneid": "4da1b188-dcd6-4ff5-b7fd-bde984055548",
				"zonename": "at-vie-1"
			  }
			]
		  }}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	zoneID := "4da1b188-dcd6-4ff5-b7fd-bde984055548"
	id := "a8a4b773-32ce-4d1c-a19b-21da055ec5c6"
	template := &Template{IsFeatured: true,
		ZoneID: zoneID,
		ID:     id,
	}

	temps, err := cs.List(template)
	if err != nil {
		t.Error(err)
	}

	if len(temps) != 1 {
		t.Fatalf("Expected one template, got %v", len(temps))
	}

	temp := temps[0].(Template)

	if temp.ZoneID != zoneID && temp.ID != id {
		t.Errorf("Wrong result")
	}
}
