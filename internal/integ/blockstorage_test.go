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

func TestBlockStorage(t *testing.T) {
	s := test.Suite{
		Zone: "ch-gva-2",
		Steps: []test.Step{
			{
				Description: "create volume",
				Command: "exo compute block-storage create my-test-volume" +
					" --size 11",
			},
			{
				Description: "check volume",
				Command:     "exo compute block-storage show my-test-volume",
				Expected: blockStorageShowOutput{
					Name:  "my-test-volume",
					Size:  "11 GiB",
					State: "detached",
				},
			},
			{
				Description: "resize volume",
				Command: "exo compute block-storage update my-test-volume" +
					" --size 12",
			},
			{
				Description: "check volume",
				Command:     "exo compute block-storage show my-test-volume",
				Expected: blockStorageShowOutput{
					Name:  "my-test-volume",
					Size:  "12 GiB",
					State: "detached",
				},
			},
			{
				Description: "clean up volume",
				Command: "exo compute block-storage delete my-test-volume" +
					" --force",
			},
		},
		T: t,
	}

	s.Run()
}
