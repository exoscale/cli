package egoscale

import (
	"encoding/json"
	"net"
	"strings"
	"testing"
)

func testClientAPIName(t *testing.T) {
	cs := NewClient("ENDPOINT", "KEY", "SECRET")
	req := &ListAPIs{}
	if cs.APIName(req) != req.name() {
		t.Errorf("APIName is wrong")
	}
}

func TestClientSyncDelete(t *testing.T) {
	respSuccessString := response{200, `
{"deleteresponse": {
	"success": "true"
}}`}
	respSuccessBool := response{200, `
{"deleteresponse": {
	"success": true
}}`}

	respError := response{431, `
{"deleteresponse": {
	"errorcode": 431,
	"cserrorcode": 9999,
	"errortext": "This is a dummy error",
	"uuidList": []
}}`}
	// those examples are not real as the API should always answer
	// with the above kind of response
	respErrorString := response{200, `
{"deleteresponse": {
	"success": "false",
	"displaytext": "herp derp"
}}`}
	respErrorBool := response{200, `
{"deleteresponse": {
	"success": false,
	"displaytext": "herp derp"
}}`}

	things := []Deletable{
		&SecurityGroup{ID: "test"},
		&SecurityGroup{Name: "test"},
		&SSHKeyPair{Name: "test"},
	}

	for _, thing := range things {
		ts := newServer(respSuccessString, respSuccessBool, respError, respErrorString, respErrorBool)
		defer ts.Close()

		cs := NewClient(ts.URL, "KEY", "SECRET")

		for i := 0; i < 2; i++ {
			if err := cs.Delete(thing); err != nil {
				t.Errorf("Deletion of %#v. Err: %s", thing, err)
			}
		}

		for i := 0; i < 3; i++ {
			if err := cs.Delete(thing); err == nil {
				t.Errorf("Deletion of %v an error was expected", thing)
			}
		}
	}
}

func TestClientAsyncDelete(t *testing.T) {
	resp := response{200, `
{"deleteresponse": {
	"jobid": "1",
	"jobresult": {
		"success": "true"
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
	things := []Deletable{
		&AffinityGroup{},
		&SecurityGroup{},
		&SSHKeyPair{},
		&VirtualMachine{},
	}

	for _, thing := range things {
		ts := newServer()
		defer ts.Close()

		cs := NewClient(ts.URL, "KEY", "SECRET")

		if err := cs.Delete(thing); err == nil {
			t.Errorf("Deletion of %#v. Should have failed", thing)
		}
	}
}

func TestClientGetNone(t *testing.T) {
	resp := response{200, `{"listfooresponse": {}}`}
	respError := response{400, `{"listfooresponse": {
		"cserrorcode": 9999,
		"errorcode": 431,
		"errortext": "Unable to execute API command due to invalid value.",
		"uuidList": []
	}}`}

	things := []Gettable{
		&Zone{ID: "1"},
		&Zone{Name: "test zone"},
		&IPAddress{ID: "1"},
		&IPAddress{IPAddress: net.ParseIP("127.0.0.1")},
		&SSHKeyPair{Name: "1"},
		&SSHKeyPair{Fingerprint: "test ssh keypair"},
		&AffinityGroup{ID: "1"},
		&AffinityGroup{Name: "test affinity group"},
		&SecurityGroup{ID: "1"},
		&SecurityGroup{Name: "test affinity group"},
		&VirtualMachine{ID: "1"},
		&Volume{ID: "1"},
		&Template{ID: "1", IsFeatured: true},
	}

	for _, thing := range things {
		ts := newServer(resp, respError)
		defer ts.Close()

		cs := NewClient(ts.URL, "KEY", "SECRET")

		for _, text := range []string{"not found", "due to invalid value"} {
			err := cs.Get(thing)
			if err == nil {
				t.Errorf("An error was expected")
				continue
			}

			e, ok := err.(*ErrorResponse)
			if !ok {
				t.Errorf("ErrorResponse was expected, got %t", err)
				continue
			}

			if !strings.Contains(e.ErrorText, text) {
				t.Errorf("Bad error test, got %q", e.ErrorText)
			}
		}
	}
}

func TestClientGetTooMany(t *testing.T) {
	resp := response{200, `{"listfooresponse": {
		"count": 2,
		"affinitygroup": [{}, {}],
		"publicipaddress": [{}, {}],
		"securitygroup": [{}, {}],
		"sshkeypair": [{}, {}],
		"virtualmachine": [{}, {}],
		"volume": [{}, {}],
		"zone": [{}, {}],
		"template": [{}, {}]
	}}`}

	things := []Gettable{
		&Zone{ID: "1"},
		&Zone{Name: "test zone"},
		&IPAddress{ID: "1"},
		&IPAddress{IPAddress: net.ParseIP("127.0.0.1")},
		&SSHKeyPair{Name: "1"},
		&SSHKeyPair{Fingerprint: "test ssh keypair"},
		&AffinityGroup{ID: "1"},
		&AffinityGroup{Name: "test affinity group"},
		&SecurityGroup{ID: "1"},
		&SecurityGroup{Name: "test affinity group"},
		&VirtualMachine{ID: "1"},
		&Volume{ID: "1"},
		&Template{ID: "1", IsFeatured: true},
	}

	for _, thing := range things {
		ts := newServer(resp)
		defer ts.Close()

		cs := NewClient(ts.URL, "KEY", "SECRET")

		// Too many
		err := cs.Get(thing)

		if err == nil {
			t.Errorf("An error was expected")
		}

		if !strings.HasPrefix(err.Error(), "More than one") {
			t.Errorf("Bad error %s", err)
		}
	}
}

func TestBooleanResponse(t *testing.T) {
	body := `{"success": true, "displaytext": "yay!"}`
	response := new(booleanResponse)

	err := json.Unmarshal([]byte(body), response)

	if err != nil {
		t.Fatalf("This shouldn't break")
	}

	success, _ := response.IsSuccess()
	if !success {
		t.Errorf("A success was expected")
	}

	if response.DisplayText != "yay!" {
		t.Errorf("DisplayText doesn't match %q", response.DisplayText)
	}
}

func TestBooleanResponseString(t *testing.T) {
	body := `{"success": "true"}`
	response := new(booleanResponse)

	err := json.Unmarshal([]byte(body), response)

	if err != nil {
		t.Fatalf("This shouldn't break")
	}

	success, _ := response.IsSuccess()
	if !success {
		t.Errorf("A success was expected")
	}
}

func TestBooleanResponseEmpty(t *testing.T) {
	body := `{}`
	response := new(booleanResponse)

	err := json.Unmarshal([]byte(body), response)

	if err != nil {
		t.Fatalf("This shouldn't break")
	}

	err = response.Error()
	if err == nil {
		t.Errorf("The booleanResponse is not a valid one")
	}
}

func TestBooleanResponseInvalid(t *testing.T) {
	body := `{"success": 42}`
	response := new(booleanResponse)

	err := json.Unmarshal([]byte(body), response)

	if err != nil {
		t.Fatalf("This shouldn't break")
	}

	err = response.Error()
	if err == nil {
		t.Errorf("An error was expected")
	}
}
