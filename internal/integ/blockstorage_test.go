package integ_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/exoscale/cli/internal/integ/test"
)

type blockStorageShowOutput struct {
	Name   string            `json:"name"`
	Zone   string            `json:"zone"`
	Size   string            `json:"size"`
	Labels map[string]string `json:"labels"`
	State  string            `json:"state"`
}

func TestBlockStorage(t *testing.T) {
	params := struct {
		VolumeName    string
		NewVolumeName string
	}{
		VolumeName:    fmt.Sprintf("test-vol-name-%d", rand.Int()),
		NewVolumeName: fmt.Sprintf("test-vol-name-%d", rand.Int()),
	}

	s := test.Suite{
		Zone:       "ch-gva-2",
		Parameters: params,
		Steps: []test.Step{
			{
				Description: "create volume",
				Command: "exo compute block-storage create {{.VolumeName}}" +
					" --size 11" +
					" --label foo1=bar1,foo2=bar2",
			},
			{
				Description: "check volume",
				Command:     "exo compute block-storage show {{.VolumeName}}",
				Expected: blockStorageShowOutput{
					Name:  params.VolumeName,
					Size:  "11 GiB",
					State: "detached",
					Labels: map[string]string{
						"foo1": "bar1",
						"foo2": "bar2",
					},
				},
			},
			{
				Description: "resize volume and change name",
				Command: "exo compute block-storage update {{.VolumeName}}" +
					" --size 12" +
					" --rename {{.NewVolumeName}}",
			},
			{
				Description: "check volume",
				Command:     "exo compute block-storage show {{.NewVolumeName}}",
				Expected: blockStorageShowOutput{
					Name:  params.NewVolumeName,
					Size:  "12 GiB",
					State: "detached",
					Labels: map[string]string{
						"foo1": "bar1",
						"foo2": "bar2",
					},
				},
			},
			{
				Description: "update volume labels",
				Command: "exo compute block-storage update {{.NewVolumeName}}" +
					" --label foo3=bar3",
			},
			{
				Description: "check volume",
				Command:     "exo compute block-storage show {{.NewVolumeName}}",
				Expected: blockStorageShowOutput{
					Name:  params.NewVolumeName,
					Size:  "12 GiB",
					State: "detached",
					Labels: map[string]string{
						"foo3": "bar3",
					},
				},
			},
			{
				Description: "clean up volume",
				Command: "exo compute block-storage delete {{.NewVolumeName}}" +
					" --force",
			},
		},
		T: t,
	}

	s.Run()
}
