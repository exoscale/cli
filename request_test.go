package egoscale

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

const (
	jsonContentType = "application/json"
)

func TestRequest(t *testing.T) {
	params := url.Values{}
	params.Set("command", "listApis")
	params.Set("apikey", "KEY")
	params.Set("name", "dummy")
	params.Set("response", "json")
	ts := newGetServer(params, jsonContentType, `
{
	"listapisresponse": {
		"api": [{
			"name": "dummy",
			"description": "this is a test",
			"isasync": false,
			"since": "4.4",
			"type": "list",
			"name": "listDummies",
			"params": [],
			"related": "",
			"response": []
		}],
		"count": 1
	}
}
	`)
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	req := &ListAPIs{
		Name: "dummy",
	}
	resp, err := cs.Request(req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	apis := resp.(*ListAPIsResponse)
	if apis.Count != 1 {
		t.Errorf("Expected exactly one API, got %d", apis.Count)
	}
}

func TestRequestSignatureFailure(t *testing.T) {
	ts := newServer(response{401, jsonContentType, `
{"createsshkeypairresponse" : {
	"uuidList":[],
	"errorcode":401,
	"errortext":"unable to verify usercredentials and/or request signature"
}}
	`})
	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	req := &CreateSSHKeyPair{
		Name: "123",
	}

	if _, err := cs.Request(req); err == nil {
		t.Errorf("This should have failed?")
		r, ok := err.(*ErrorResponse)
		if !ok {
			t.Errorf("A CloudStack error was expected, got %v", err)
		}
		if r.ErrorCode != Unauthorized {
			t.Errorf("Unauthorized error was expected")
		}
	}
}

func TestBooleanAsyncRequest(t *testing.T) {
	ts := newServer(response{200, jsonContentType, `
{
	"expungevirtualmachineresponse": {
		"jobid": "1",
		"jobresult": {},
		"jobstatus": 0
	}
}
	`}, response{200, jsonContentType, `
{
	"queryasyncjobresultresponse": {
		"accountid": "1",
		"cmd": "expunge",
		"created": "2018-04-03T22:40:04+0200",
		"jobid": "1",
		"jobprocstatus": 0,
		"jobresult": {
			"success": true
		},
		"jobresultcode": 0,
		"jobresulttype": "object",
		"jobstatus": 1,
		"userid": "1"
	}
}
	`})
	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	req := &ExpungeVirtualMachine{
		ID: "123",
	}
	if err := cs.BooleanRequest(req); err != nil {
		t.Error(err)
	}
}

func TestBooleanAsyncRequestWithContext(t *testing.T) {
	ts := newServer(response{200, jsonContentType, `
{
	"expungevirtualmachineresponse": {
		"jobid": "1",
		"jobresult": {},
		"jobstatus": 0
	}
}
	`}, response{200, jsonContentType, `
{
	"queryasyncjobresultresponse": {
		"accountid": "1",
		"cmd": "expunge",
		"created": "2018-04-03T22:40:04+0200",
		"jobid": "1",
		"jobprocstatus": 0,
		"jobresult": {
			"success": true
		},
		"jobresultcode": 0,
		"jobresulttype": "object",
		"jobstatus": 1,
		"userid": "1"
	}
}
	`})
	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	req := &ExpungeVirtualMachine{
		ID: "123",
	}

	// WithContext
	if err := cs.BooleanRequestWithContext(context.Background(), req); err != nil {
		t.Error(err)
	}
}

func TestBooleanRequestTimeout(t *testing.T) {
	ts := newSleepyServer(time.Second, 200, jsonContentType, `
{
	"expungevirtualmachine": {
		"jobid": "1",
		"jobresult": {
			"success": false
		},
		"jobstatus": 0
	}
}
	`)
	defer ts.Close()
	done := make(chan bool)

	go func() {
		cs := NewClientWithTimeout(ts.URL, "TOKEN", "SECRET", time.Millisecond)

		req := &ExpungeVirtualMachine{
			ID: "123",
		}
		err := cs.BooleanRequest(req)

		if err == nil {
			t.Error("An error was expected")
		}

		// We expect the HTTP Client to timeout
		msg := err.Error()
		if !strings.HasPrefix(msg, "Get") {
			t.Errorf("Unexpected error message: %s", err.Error())
		}

		done <- true
	}()

	<-done
}

func TestSyncRequestWithoutContext(t *testing.T) {

	ts := newServer(
		response{200, jsonContentType, `{
	"deployvirtualmachineresponse": {
		"jobid": "42",
		"jobresult": {},
		"jobstatus": 0
	}
}`},
	)

	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	req := &DeployVirtualMachine{
		Name:              "test",
		ServiceOfferingID: "71004023-bb72-4a97-b1e9-bc66dfce9470",
		ZoneID:            "1128bd56-b4d9-4ac6-a7b9-c715b187ce11",
		TemplateID:        "78c2cbe6-8e11-4722-b01f-bf06f4e28108",
	}

	// WithContext
	resp, err := cs.SyncRequest(req)
	if err != nil {
		t.Error(err)
	}
	result, ok := resp.(*AsyncJobResult)
	if !ok {
		t.Error("wrong type")
	}

	if result.JobID != "42" {
		t.Errorf("wrong job id, expected 42, got %s", result.JobID)
	}
}

func TestAsyncRequestWithoutContext(t *testing.T) {

	ts := newServer(
		response{200, jsonContentType, `{
	"deployvirtualmachineresponse": {
		"jobid": "1",
		"jobresult": {},
		"jobstatus": 0
	}
}`},
		response{200, jsonContentType, `{
	"queryasyncjobresultresponse": {
		"jobid": "1",
		"jobresult": {
			"virtualmachine": {
				"id": "f344b886-2a8b-4d2c-9662-1f18e5cdde6f",
				"serviceofferingid": "71004023-bb72-4a97-b1e9-bc66dfce9470",
				"templateid": "78c2cbe6-8e11-4722-b01f-bf06f4e28108",
				"zoneid": "1128bd56-b4d9-4ac6-a7b9-c715b187ce11",
				"jobid": "220504ac-b9e7-4fee-b402-47b3c4155fdb"
			}
		},
		"jobstatus": 1
	}
}`},
	)

	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	req := &DeployVirtualMachine{
		Name:              "test",
		ServiceOfferingID: "71004023-bb72-4a97-b1e9-bc66dfce9470",
		ZoneID:            "1128bd56-b4d9-4ac6-a7b9-c715b187ce11",
		TemplateID:        "78c2cbe6-8e11-4722-b01f-bf06f4e28108",
	}

	resp := &VirtualMachine{}

	// WithContext
	cs.AsyncRequest(req, func(j *AsyncJobResult, err error) bool {
		if err != nil {
			t.Error(err)
			return false
		}

		if j.JobStatus == Success {
			if r := j.Result(resp); r != nil {
				t.Error(r)
			}
			return false
		}
		return true
	})

	if resp.ServiceOfferingID != "71004023-bb72-4a97-b1e9-bc66dfce9470" {
		t.Errorf("Expected ServiceOfferingID %q, got %q", "71004023-bb72-4a97-b1e9-bc66dfce9470", resp.ServiceOfferingID)
	}
}

func TestAsyncRequestWithoutContextFailure(t *testing.T) {
	ts := newServer(
		response{200, jsonContentType, `{
	"deployvirtualmachineresponse": {
		"jobid": "1",
		"jobresult": {},
		"jobstatus": 0
	}
}`},
		response{200, jsonContentType, `{
	"queryasyncjobresultresponse": {
		"jobid": "1",
		"jobresult": {
			"virtualmachine": []
		},
		"jobstatus": 1
	}
}`},
	)

	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	req := &DeployVirtualMachine{
		Name:              "test",
		ServiceOfferingID: "71004023-bb72-4a97-b1e9-bc66dfce9470",
		ZoneID:            "1128bd56-b4d9-4ac6-a7b9-c715b187ce11",
		TemplateID:        "78c2cbe6-8e11-4722-b01f-bf06f4e28108"}

	resp := &VirtualMachine{}

	// WithContext
	cs.AsyncRequest(req, func(j *AsyncJobResult, err error) bool {
		if err != nil {
			t.Fatal(err)
		}

		if j.JobStatus == Success {

			if r := j.Result(resp); r != nil {
				return false
			}
			t.Errorf("Expected an error, got <nil>")
		}
		return true
	})
}

func TestAsyncRequestWithoutContextFailureNext(t *testing.T) {

	ts := newServer(
		response{200, jsonContentType, `{
	"deployvirtualmachineresponse: {
		"jobid": "1",
		"jobresult": {},
		"jobstatus": 0
	}
}`},
	)

	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	req := &DeployVirtualMachine{
		Name:              "test",
		ServiceOfferingID: "71004023-bb72-4a97-b1e9-bc66dfce9470",
		ZoneID:            "1128bd56-b4d9-4ac6-a7b9-c715b187ce11",
		TemplateID:        "78c2cbe6-8e11-4722-b01f-bf06f4e28108",
	}

	cs.AsyncRequest(req, func(j *AsyncJobResult, err error) bool {
		return err == nil
	})
}

func TestAsyncRequestWithoutContextFailureNextNext(t *testing.T) {

	ts := newServer(
		response{200, jsonContentType, `{
	"deployvirtualmachineresponse": {
		"jobid": "1",
		"jobresult": {
			"virtualmachine": {}
		},
		"jobstatus": 2
	}
}`},
		response{200, jsonContentType, `{
	"queryasyncjobresultresponse": {
		"jobid": "1",
		"jobresult": {},
		"jobstatus": 0
	}
}`},
		response{200, jsonContentType, `{
	"queryasyncjobresultresponse": {
		"jobid": "1",
		"jobresult": [],
		"jobstatus": 1
	}
}`},
	)
	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	req := &DeployVirtualMachine{
		Name:              "test",
		ServiceOfferingID: "71004023-bb72-4a97-b1e9-bc66dfce9470",
		ZoneID:            "1128bd56-b4d9-4ac6-a7b9-c715b187ce11",
		TemplateID:        "78c2cbe6-8e11-4722-b01f-bf06f4e28108"}

	resp := &VirtualMachine{}

	cs.AsyncRequest(req, func(j *AsyncJobResult, err error) bool {
		if err != nil {
			t.Fatal(err)
		}

		if j.JobStatus == Success {

			j.JobStatus = Failure
			if r := j.Result(resp); r != nil {
				return false
			}
			t.Errorf("Expected an error, got <nil>")
		}
		return true
	})
}

func TestBooleanRequestWithContext(t *testing.T) {
	ts := newSleepyServer(time.Second, 200, jsonContentType, `
{
	"expungevirtualmachine": {
		"jobid": "1",
		"jobresult": {
			"success": false
		},
		"jobstatus": 0
	}
}
	`)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	done := make(chan bool)

	go func() {
		cs := NewClient(ts.URL, "TOKEN", "SECRET")
		req := &ExpungeVirtualMachine{
			ID: "123",
		}

		err := cs.BooleanRequestWithContext(ctx, req)

		if err == nil {
			t.Error("An error was expected")
		}

		// We expect the context to timeout
		msg := err.Error()
		if !strings.HasPrefix(msg, "Get") {
			t.Errorf("Unexpected error message: %s", err.Error())
		}

		done <- true
	}()

	<-done
}

func TestRequestWithContextTimeoutPost(t *testing.T) {
	ts := newSleepyServer(time.Second, 200, jsonContentType, `
{
	"deployvirtualmachineresponse": {
		"jobid": "1",
		"jobresult": {
			"success": false
		},
		"jobstatus": 0
	}
}
	`)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	done := make(chan bool)

	userData := make([]byte, 1<<11)
	_, err := rand.Read(userData)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		cs := NewClient(ts.URL, "TOKEN", "SECRET")
		req := &DeployVirtualMachine{
			ServiceOfferingID: "test",
			TemplateID:        "test",
			UserData:          base64.StdEncoding.EncodeToString(userData),
			ZoneID:            "test",
		}

		_, err := cs.RequestWithContext(ctx, req)

		if err == nil {
			t.Error("An error was expected")
		}

		// We expect the context to timeout
		msg := err.Error()
		if !strings.HasPrefix(msg, "Post") {
			t.Errorf("Unexpected error message: %s", err.Error())
		}

		done <- true
	}()

	<-done
}

func TestBooleanRequestWithContextAndTimeout(t *testing.T) {
	ts := newSleepyServer(time.Second, 200, jsonContentType, `
{
	"expungevirtualmachine": {
		"jobid": "1",
		"jobresult": {
			"success": false
		},
		"jobstatus": 0
	}
}
	`)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan bool)

	go func() {
		cs := NewClientWithTimeout(ts.URL, "TOKEN", "SECRET", time.Millisecond)
		req := &ExpungeVirtualMachine{
			ID: "123",
		}
		err := cs.BooleanRequestWithContext(ctx, req)

		if err == nil {
			t.Error("An error was expected")
		}

		// We expect the client to timeout
		msg := err.Error()
		if !strings.HasPrefix(msg, "Get") || !strings.Contains(msg, "net/http: request canceled") {
			t.Errorf("Unexpected error message: %s", err.Error())
		}

		done <- true
	}()

	<-done
}

func TestWrongBodyResponse(t *testing.T) {
	ts := newServer(response{200, "text/html", `
		<html>
		<header><title>This is title</title></header>
		<body>
		Hello world
		</body>
		</html>		
	`})
	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")

	_, err := cs.Request(&ListZones{})
	if err == nil {
		t.Error("an error was expected but got nil error")
	}

	if err.Error() != fmt.Sprintf("body content-type response expected %q, got %q", jsonContentType, "text/html") {
		t.Error("body content-type error response expected")
	}
}

type response struct {
	code        int
	contentType string
	body        string
}

func newServer(responses ...response) *httptest.Server {
	i := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if i >= len(responses) {
			w.Header().Set("Content-Type", jsonContentType)
			w.WriteHeader(500)
			w.Write([]byte("{}"))
			return
		}
		w.Header().Set("Content-Type", responses[i].contentType)
		w.WriteHeader(responses[i].code)
		w.Write([]byte(responses[i].body))
		i++
	})
	return httptest.NewServer(mux)
}

func newSleepyServer(sleep time.Duration, code int, contentType, response string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(sleep)
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(code)
		w.Write([]byte(response))
	})
	return httptest.NewServer(mux)
}

func newGetServer(params url.Values, contentType, response string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		errors := make(map[string][]string)
		query := r.URL.Query()
		for k, expected := range params {
			if values, ok := query[k]; ok {
				for i, value := range values {
					e := expected[i]
					if e != value {
						if _, ok := errors[k]; !ok {
							errors[k] = make([]string, len(values))
						}
						errors[k][i] = fmt.Sprintf("%s expected %v, got %v", k, e, value)
					}
				}
			} else {
				errors[k] = make([]string, 1)
				errors[k][0] = fmt.Sprintf("%s was expected", k)
			}
		}

		if len(errors) == 0 {
			w.Header().Set("Content-Type", contentType)
			w.WriteHeader(200)
			w.Write([]byte(response))
		} else {
			w.Header().Set("Content-Type", contentType)
			w.WriteHeader(400)
			body, _ := json.Marshal(errors)
			w.Write(body)
		}
	})
	return httptest.NewServer(mux)
}
