package egoscale

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestRequest(t *testing.T) {
	params := url.Values{}
	params.Set("command", "listApis")
	params.Set("token", "TOKEN")
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

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
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

func TestBooleanAsyncRequest(t *testing.T) {
	params := url.Values{}
	params.Set("command", "expungevirtualmachine")
	params.Set("token", "TOKEN")
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
	err := cs.BooleanRequest(req)

	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestBooleanAsyncRequestTimeout(t *testing.T) {
	params := url.Values{}
	params.Set("command", "expungevirtualmachine")
	params.Set("token", "TOKEN")
	params.Set("id", "123")
	params.Set("response", "json")
	ts := newPostServer(params, `
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

	cs := NewClientWithTimeout(ts.URL, "TOKEN", "SECRET", time.Second)
	req := &ExpungeVirtualMachine{
		ID: "123",
	}
	err := cs.BooleanRequest(req)

	if err == nil {
		t.Error("An error was expected")
	}

	if err.Error() != "context deadline exceeded" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func newServer(code int, response string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Write([]byte(response))
	})
	return httptest.NewServer(mux)
}

func newPostServer(params url.Values, response string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		errors := make(map[string][]string)
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
			}
		}

		if len(errors) == 0 {
			w.WriteHeader(200)
			w.Write([]byte(response))
		} else {
			w.WriteHeader(400)
			body, _ := json.Marshal(errors)
			w.Write(body)
			log.Println(body)
		}
	})
	return httptest.NewServer(mux)
}
