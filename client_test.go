package egoscale

import (
	"testing"
)

func TestClientSyncDelete(t *testing.T) {
	resp := response{200, `
{"deleteresponse": {
	"success": "true"
}}`}
	respError := response{400, `
	{"deleteresponse": {
		"success": "false",
		"displaytext": "herp derp"
	}}`}

	things := []Deletable{
		&SecurityGroup{ID: "test"},
		&SecurityGroup{Name: "test"},
		&SSHKeyPair{Name: "test"},
	}

	for _, thing := range things {
		ts := newServer(resp, respError)
		defer ts.Close()

		cs := NewClient(ts.URL, "KEY", "SECRET")

		if err := cs.Delete(thing); err != nil {
			t.Errorf("Deletion of %#v. Err: %s", thing, err)
		}

		if err := cs.Delete(thing); err == nil {
			t.Errorf("Deletion of %#v. An error was expected", thing)
		}
	}
}

func TestClientAsyncDelete(t *testing.T) {
	resp := response{200, `
{"deleteresponse": {
	"jobid": "1",
	"jobresult": {
		"success": true
	},
	"jobstatus": 1
}}`}
	respError := response{400, `
{"deleteresponse": {
	"jobid": "1",
	"jobresult": {
		"success": false,
		"displaytext": "herp derp",
	},
	"jobstatus": 2
}}`}

	things := []Deletable{
		&AffinityGroup{ID: "affinity group id"},
		&AffinityGroup{Name: "affinity group name"},
		&IPAddress{ID: "ip address id"},
		&VirtualMachine{ID: "virtual machine id"},
	}

	for _, thing := range things {
		ts := newServer(resp, respError)
		defer ts.Close()

		cs := NewClient(ts.URL, "KEY", "SECRET")

		if err := cs.Delete(thing); err != nil {
			t.Errorf("Deletion of %#v. Err: %s", thing, err)
		}
		if err := cs.Delete(thing); err == nil {
			t.Errorf("Deletion of %#v. An error was expected", thing)
		}
	}
}

func TestClientDeleteFailure(t *testing.T) {
	ts := newServer()
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")

	things := []Deletable{
		&AffinityGroup{},
		&SecurityGroup{},
		&SSHKeyPair{},
		&VirtualMachine{},
	}

	for _, thing := range things {
		if err := cs.Delete(thing); err == nil {
			t.Errorf("Deletion of %#v. Should have failed", thing)
		}
	}
}
