package egoscale

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestClientAPIName(t *testing.T) {
	cs := NewClient("ENDPOINT", "KEY", "SECRET")
	req := &ListAPIs{}
	if cs.APIName(req) != "listApis" {
		t.Errorf("APIName is wrong, wanted listApis")
	}
	if cs.APIName(&AuthorizeSecurityGroupIngress{}) != "authorizeSecurityGroupIngress" {
		t.Errorf("APIName is wrong, wanted Ingress")
	}
	if cs.APIName(&AuthorizeSecurityGroupEgress{}) != "authorizeSecurityGroupEgress" {
		t.Errorf("APIName is wrong, wanted Egress")
	}
}

func TestClientResponse(t *testing.T) {
	cs := NewClient("ENDPOINT", "KEY", "SECRET")

	r := cs.Response(&ListAPIs{})
	switch r.(type) {
	case *ListAPIsResponse:
		// do nothing
	default:
		t.Errorf("request is wrong, got %t", r)
	}

	ar := cs.Response(&DeployVirtualMachine{})
	switch ar.(type) {
	case *VirtualMachine:
		// do nothing
	default:
		t.Errorf("asyncRequest is wrong, got %t", ar)
	}
}

func TestClientSyncDelete(t *testing.T) {
	bodySuccessString := `
{"delete%sresponse": {
	"success": "true"
}}`
	bodySuccessBool := `
{"delete%sresponse": {
	"success": true
}}`

	bodyError := `
{"delete%sresponse": {
	"errorcode": 431,
	"cserrorcode": 9999,
	"errortext": "This is a dummy error",
	"uuidList": []
}}`

	things := []struct {
		name      string
		deletable Deletable
	}{
		{"securitygroup", &SecurityGroup{ID: "test"}},
		{"securitygroup", &SecurityGroup{Name: "test"}},
		{"sshkeypair", &SSHKeyPair{Name: "test"}},
	}

	for _, thing := range things {
		ts := newServer(
			response{200, jsonContentType, fmt.Sprintf(bodySuccessString, thing.name)},
			response{200, jsonContentType, fmt.Sprintf(bodySuccessBool, thing.name)},
			response{431, jsonContentType, fmt.Sprintf(bodyError, thing.name)},
		)
		defer ts.Close()

		cs := NewClient(ts.URL, "KEY", "SECRET")

		for i := 0; i < 2; i++ {
			if err := cs.Delete(thing.deletable); err != nil {
				t.Errorf("Deletion of %#v. Err: %s", thing.deletable, err)
			}
		}

		if err := cs.Delete(thing.deletable); err == nil {
			t.Errorf("Deletion of %v an error was expected", thing.deletable)
		}
	}
}

func TestClientAsyncDelete(t *testing.T) {
	body := `
{"%sresponse": {
	"jobid": "1",
	"jobresult": {
		"success": "true"
	},
	"jobstatus": 1
}}`
	bodyError := `
{"%sresponse": {
	"jobid": "1",
	"jobresult": {
		"success": false,
		"displaytext": "herp derp",
	},
	"jobstatus": 2
}}`

	things := []struct {
		name      string
		deletable Deletable
	}{
		{"deleteaffinitygroup", &AffinityGroup{ID: "affinity group id"}},
		{"deleteaffinitygroup", &AffinityGroup{Name: "affinity group name"}},
		{"disassociateipaddress", &IPAddress{ID: "ip address id"}},
		{"destroyvirtualmachine", &VirtualMachine{ID: "virtual machine id"}},
	}

	for _, thing := range things {
		ts := newServer(
			response{200, jsonContentType, fmt.Sprintf(body, thing.name)},
			response{400, jsonContentType, fmt.Sprintf(bodyError, thing.name)},
		)
		defer ts.Close()

		cs := NewClient(ts.URL, "KEY", "SECRET")

		if err := cs.Delete(thing.deletable); err != nil {
			t.Errorf("Deletion of %#v. Err: %s", thing, err)
		}
		if err := cs.Delete(thing.deletable); err == nil {
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
		&IPAddress{},
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

func TestClientGetFailure(t *testing.T) {
	things := []Gettable{
		nil,
		&AffinityGroup{},
		&SecurityGroup{},
		&SSHKeyPair{},
		&VirtualMachine{},
		&IPAddress{},
		&Account{},
	}

	for _, thing := range things {
		ts := newServer()
		defer ts.Close()

		cs := NewClient(ts.URL, "KEY", "SECRET")

		if err := cs.Get(thing); err == nil {
			t.Errorf("Get of %#v. Should have failed", thing)
		}
	}
}

func TestClientGetNone(t *testing.T) {
	body := `{"list%sresponse": {}}`
	bodyError := `{"errorresponse": {
		"cserrorcode": 9999,
		"errorcode": 431,
		"errortext": "Unable to execute API command due to invalid value.",
		"uuidList": []
	}}`

	things := []struct {
		name     string
		gettable Gettable
	}{
		{"zones", &Zone{ID: "1"}},
		{"zones", &Zone{Name: "test zone"}},
		{"publicipaddresses", &IPAddress{ID: "1"}},
		{"publicipaddresses", &IPAddress{IPAddress: net.ParseIP("127.0.0.1")}},
		{"sshkeypairs", &SSHKeyPair{Name: "1"}},
		{"sshkeypairs", &SSHKeyPair{Fingerprint: "test ssh keypair"}},
		{"affinitygroups", &AffinityGroup{ID: "1"}},
		{"affinitygroups", &AffinityGroup{Name: "test affinity group"}},
		{"securitygroups", &SecurityGroup{ID: "1"}},
		{"securitygroups", &SecurityGroup{Name: "test affinity group"}},
		{"virtualmachines", &VirtualMachine{ID: "1"}},
		{"volumes", &Volume{ID: "1"}},
		{"templates", &Template{ID: "1", IsFeatured: true}},
		{"serviceofferings", &ServiceOffering{ID: "1"}},
		{"accounts", &Account{}},
	}

	for _, thing := range things {
		ts := newServer(
			response{200, jsonContentType, fmt.Sprintf(body, thing.name)},
			response{431, jsonContentType, bodyError},
		)
		defer ts.Close()

		cs := NewClient(ts.URL, "KEY", "SECRET")

		for _, text := range []string{"not found", "due to invalid value"} {
			err := cs.Get(thing.gettable)
			if err == nil {
				t.Error("an error was expected")
				continue
			}

			e, ok := err.(*ErrorResponse)
			if !ok {
				t.Errorf("an ErrorResponse was expected, got %T", err)
				continue
			}

			if !strings.Contains(e.ErrorText, text) {
				t.Errorf("bad error test, got %q", e.ErrorText)
			}
		}
	}
}

func TestClientGetTooMany(t *testing.T) {
	body := `
	{"list%sresponse": {
		"count": 2,
		"affinitygroup": [{}, {}],
		"publicipaddress": [{}, {}],
		"securitygroup": [{}, {}],
		"sshkeypair": [{}, {}],
		"virtualmachine": [{}, {}],
		"volume": [{}, {}],
		"zone": [{}, {}],
		"template": [{}, {}],
		"serviceoffering": [{}, {}],
		"account": [{}, {}]
	}}`

	things := []struct {
		name     string
		gettable Gettable
	}{
		{"zones", &Zone{ID: "1"}},
		{"zones", &Zone{Name: "test zone"}},
		{"publicipaddresses", &IPAddress{ID: "1"}},
		{"publicipaddresses", &IPAddress{IPAddress: net.ParseIP("127.0.0.1")}},
		{"sshkeypairs", &SSHKeyPair{Name: "1"}},
		{"sshkeypairs", &SSHKeyPair{Fingerprint: "test ssh keypair"}},
		{"affinitygroups", &AffinityGroup{ID: "1"}},
		{"affinitygroups", &AffinityGroup{Name: "test affinity group"}},
		{"securitygroups", &SecurityGroup{ID: "1"}},
		{"securitygroups", &SecurityGroup{Name: "test affinity group"}},
		{"virtualmachines", &VirtualMachine{ID: "1"}},
		{"volumes", &Volume{ID: "1"}},
		{"templates", &Template{ID: "1", IsFeatured: true}},
		{"serviceofferings", &ServiceOffering{ID: "1"}},
		{"accounts", &Account{}},
	}

	for _, thing := range things {
		resp := response{200, jsonContentType, fmt.Sprintf(body, thing.name)}
		ts := newServer(resp)
		defer ts.Close()

		cs := NewClient(ts.URL, "KEY", "SECRET")

		// Too many
		err := cs.Get(thing.gettable)

		if err == nil {
			t.Errorf("an error was expected")
		}

		if !strings.HasPrefix(err.Error(), "more than one") {
			t.Errorf("bad error %s", err)
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
