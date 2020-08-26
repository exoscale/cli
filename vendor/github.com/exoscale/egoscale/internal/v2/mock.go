package v2

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
	*httpmock.MockTransport
	ClientWithResponsesInterface
}

func NewMockClient() *MockClient {
	var client MockClient

	client.MockTransport = httpmock.NewMockTransport()

	return &client
}

func (c *MockClient) Do(req *http.Request) (*http.Response, error) {
	var hc = http.Client{Transport: c.MockTransport}

	return hc.Do(req)
}
