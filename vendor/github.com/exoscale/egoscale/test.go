package egoscale

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

var testZone = "ch-gva-2"

func testUnmarshalJSONRequestBody(t *testing.T, req *http.Request, v interface{}) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("error reading request body: %s", err)
	}
	if err = json.Unmarshal(data, v); err != nil {
		t.Fatalf("error while unmarshalling JSON body: %s", err)
	}
}
