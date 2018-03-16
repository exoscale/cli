package egoscale

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestRequest(t *testing.T) {
	params := url.Values{}
	params.Set("command", "listApis")
	params.Set("apikey", "KEY")
	params.Set("name", "dummy")
	params.Set("response", "json")
	ts := newPostServer(params, `
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
	ts := newServer(401, `
{"createsshkeypairresponse" : {
	"uuidList":[],
	"errorcode":401,
	"errortext":"unable to verify usercredentials and/or request signature"
}}
	`)
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
	params := url.Values{}
	params.Set("command", "expungeVirtualMachine")
	params.Set("apikey", "TOKEN")
	params.Set("id", "123")
	params.Set("response", "json")
	ts := newPostServer(params, `
{
	"expungevirtualmarchine": {
		"jobid": "1",
		"jobresult": {
			"success": true,
			"displaytext": "good job!"
		},
		"jobstatus": 1
	}
}
	`)
	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	req := &ExpungeVirtualMachine{
		ID: "123",
	}
	if err := cs.BooleanRequest(req); err != nil {
		t.Error(err)
	}

	// WithContext
	if err := cs.BooleanRequestWithContext(context.Background(), req); err != nil {
		t.Error(err)
	}
}

func TestBooleanRequestTimeout(t *testing.T) {
	ts := newSleepyServer(time.Second, 200, `
{
	"expungevirtualmarchine": {
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
		if !strings.HasPrefix(msg, "Post") {
			t.Errorf("Unexpected error message: %s", err.Error())
		}

		done <- true
	}()

	<-done
}

func TestBooleanRequestWithContext(t *testing.T) {
	ts := newSleepyServer(time.Second, 200, `
{
	"expungevirtualmarchine": {
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
		if !strings.HasPrefix(msg, "Post") {
			t.Errorf("Unexpected error message: %s", err.Error())
		}

		done <- true
	}()

	<-done
}

func TestBooleanRequestWithContextAndTimeout(t *testing.T) {
	ts := newSleepyServer(time.Second, 200, `
{
	"expungevirtualmarchine": {
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
		if !strings.HasPrefix(msg, "Post") || !strings.Contains(msg, "net/http: request canceled") {
			t.Errorf("Unexpected error message: %s", err.Error())
		}

		done <- true
	}()

	<-done
}

func newServer(code int, response string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Write([]byte(response))
	})
	return httptest.NewServer(mux)
}

func newSleepyServer(sleep time.Duration, code int, response string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(sleep)
		w.WriteHeader(code)
		w.Write([]byte(response))
	})
	return httptest.NewServer(mux)
}

func newPostServer(params url.Values, response string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		errors := make(map[string][]string)
		if r.ParseForm() != nil {
			w.WriteHeader(500)
			w.Write([]byte("Cannot parse the form"))
			return
		}
		for k, expected := range params {
			if values, ok := (r.PostForm)[k]; ok {
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
			w.WriteHeader(200)
			w.Write([]byte(response))
		} else {
			w.WriteHeader(400)
			body, _ := json.Marshal(errors)
			w.Write(body)
		}
	})
	return httptest.NewServer(mux)
}
