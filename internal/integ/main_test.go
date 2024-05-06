package integ_test

import (
	"testing"

	"github.com/exoscale/cli/internal/integ/test"
)

// TODO make test binary configurable

func TestMain(m *testing.M) {
	m.Run()
}

type blockStorageListItemOutput struct {
	Name  string `json:"name"`
	Zone  string `json:"zone"`
	Size  string `json:"size"`
	State string `json:"state"`
}

func TestBlockStorageCreate(t *testing.T) {
	cases := []test.Case{
		{
			Command: "exo -z ch-gva-2 -O json c bs list",
			Expected: []blockStorageListItemOutput{
				{
					Name:  "my-existing-volume",
					Size:  "11 GiB",
					State: "detached",
				},
			},
		},
	}

	s := test.Suite{
		Cases: cases,
		T:     t,
	}
	s.Run()
}
