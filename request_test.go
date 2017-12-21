package egoscale

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestPrepareValues(t *testing.T) {
	type tag struct {
		Name      string `json:"name"`
		IsVisible bool   `json:"isvisible,omitempty"`
	}

	profile := struct {
		IgnoreMe string
		Zone     string  `json:"myzone,omitempty"`
		Name     string  `json:"name"`
		Id       int     `json:"id"`
		Uid      uint    `json:"uid"`
		Num      float64 `json:"num"`
		Bytes    []byte  `json:"bytes"`
		Tags     []*tag  `json:"tags,omitempty"`
	}{
		IgnoreMe: "bar",
		Name:     "world",
		Id:       1,
		Uid:      uint(2),
		Num:      3.14,
		Bytes:    []byte("exo"),
		Tags: []*tag{
			{Name: "foo"},
			{Name: "bar", IsVisible: false},
		},
	}

	params := url.Values{}
	err := prepareValues("", &params, profile)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := params["myzone"]; ok {
		t.Errorf("myzone params shouldn't be set, got %v", params.Get("myzone"))
	}

	if params.Get("name") != "world" {
		t.Errorf("name params wasn't properly set, got %v", params.Get("name"))
	}

	if params.Get("bytes") != "ZXhv" {
		t.Errorf("bytes params wasn't properly encoded in base 64, got %v", params.Get("bytes"))
	}

	if _, ok := params["ignoreme"]; ok {
		t.Errorf("IgnoreMe key was set")
	}

	v := params.Get("tags[0].name")
	if v != "foo" {
		t.Errorf("expected tags to be serialized as foo, got %#v", v)
	}
}

func TestPrepareValuesStringRequired(t *testing.T) {
	profile := struct {
		RequiredField string `json:"requiredfield"`
	}{}

	params := url.Values{}
	err := prepareValues("", &params, &profile)
	if err == nil {
		t.Errorf("It should have failed")
	}
}

func TestPrepareValuesBoolRequired(t *testing.T) {
	profile := struct {
		RequiredField bool `json:"requiredfield"`
	}{}

	params := url.Values{}
	err := prepareValues("", &params, &profile)
	if err != nil {
		t.Fatal(nil)
	}
	if params.Get("requiredfield") != "false" {
		t.Errorf("bool params wasn't set to false (default value)")
	}
}

func TestPrepareValuesIntRequired(t *testing.T) {
	profile := struct {
		RequiredField int64 `json:"requiredfield"`
	}{}

	params := url.Values{}
	err := prepareValues("", &params, &profile)
	if err == nil {
		t.Errorf("It should have failed")
	}
}

func TestPrepareValuesUintRequired(t *testing.T) {
	profile := struct {
		RequiredField uint64 `json:"requiredfield"`
	}{}

	params := url.Values{}
	err := prepareValues("", &params, &profile)
	if err == nil {
		t.Errorf("It should have failed")
	}
}

func TestPrepareValuesBytesRequired(t *testing.T) {
	profile := struct {
		RequiredField []byte `json:"requiredfield"`
	}{}

	params := url.Values{}
	err := prepareValues("", &params, &profile)
	if err == nil {
		t.Errorf("It should have failed")
	}
}

func TestPrepareValuesSliceString(t *testing.T) {
	profile := struct {
		RequiredField []string `json:"requiredfield"`
	}{}

	params := url.Values{}
	err := prepareValues("", &params, &profile)
	if err == nil {
		t.Errorf("It should have failed")
	}
}

func TestBooleanRequest(t *testing.T) {
	params := url.Values{}
	params.Set("command", "destroyVirtualMachine")
	params.Set("token", "TOKEN")
	params.Set("id", "123")
	params.Set("response", "json")
	ts := newPostServer(params, `
{
	"destroyvirtualmachine": {
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
	req := &DestroyVirtualMachineRequest{
		Id: "123",
	}
	err := cs.BooleanAsyncRequest(req, AsyncInfo{})

	if err != nil {
		t.Errorf(err.Error())
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

		log.Printf("len %d", len(errors))
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
