package integ_test

import (
	"testing"

	"github.com/exoscale/cli/internal/integ/test"
)

type blockStorageShowOutput struct {
	Name  string `json:"name"`
	Zone  string `json:"zone"`
	Size  string `json:"size"`
	State string `json:"state"`
}

func TestBlockStorageCreateDelete(t *testing.T) {
	s := test.Suite{
		Zone: "ch-gva-2",
		Cases: []test.Case{
			{
				Command: "exo compute block-storage create my-test-volume" +
					" --size 12",
			},
			{
				Command: "exo compute block-storage show my-test-volume",
				Expected: blockStorageShowOutput{
					Name:  "my-test-volume",
					Size:  "12 GiB",
					State: "detached",
				},
			},
			{
				Command: "exo compute block-storage delete my-test-volume" +
					" --force",
			},
		},
		T: t,
	}

	s.Run()
}
