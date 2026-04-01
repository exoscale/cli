//go:build integration_api
// +build integration_api

package integration_with_api_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/exoscale/cli/internal/integ"
)

type blockStorageShowOutput struct {
	Name   string            `json:"name"`
	Zone   string            `json:"zone"`
	Size   string            `json:"size"`
	Labels map[string]string `json:"labels"`
	State  string            `json:"state"`
}

type blockStorageSnapshotShowOutput struct {
	Name   string            `json:"name"`
	Size   string            `json:"size"`
	State  string            `json:"state"`
	Labels map[string]string `json:"labels"`
}

func TestBlockStorage(t *testing.T) {
	params := struct {
		VolumeName      string
		NewVolumeName   string
		SnapshotName    string
		NewSnapshotName string
	}{
		VolumeName:      fmt.Sprintf("test-vol-name-%d", rand.Int()),
		NewVolumeName:   fmt.Sprintf("test-vol-name-%d-renamed", rand.Int()),
		SnapshotName:    fmt.Sprintf("test-snap-name-%d", rand.Int()),
		NewSnapshotName: fmt.Sprintf("test-snap-name-%d-renamed", rand.Int()),
	}

	s := integ.Suite{
		Zone:       "ch-gva-2",
		Parameters: params,
		Steps: []integ.Step{
			{
				Description: "create volume",
				Command: "exo compute block-storage create {{.VolumeName}}" +
					" --size 11" +
					" --label foo1=bar1,foo2=bar2",
			},
			{
				Description: "check created volume",
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
				Description: "check resized volume",
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
				Description: "create snapshot",
				Command: "exo compute block-storage snapshot create {{.NewVolumeName}}" +
					" --name {{.SnapshotName}}" +
					" --label ping=pong,key=val",
			},
			{
				Description: "check created snapshot",
				Command:     "exo compute block-storage snapshot show {{.SnapshotName}}",
				Expected: blockStorageSnapshotShowOutput{
					Name:  params.SnapshotName,
					Size:  "0 GiB",
					State: "created",
					Labels: map[string]string{
						"ping": "pong",
						"key":  "val",
					},
				},
			},
			{
				Description: "update snapshot name",
				Command: "exo compute block-storage snapshot update {{.SnapshotName}}" +
					" --rename {{.NewSnapshotName}}",
				Expected: blockStorageSnapshotShowOutput{
					Name:  params.NewSnapshotName,
					Size:  "0 GiB",
					State: "created",
					Labels: map[string]string{
						"ping": "pong",
						"key":  "val",
					},
				},
			},
			{
				Description: "update snapshot labels",
				Command: "exo compute block-storage snapshot update {{.NewSnapshotName}}" +
					" --label new=label",
				Expected: blockStorageSnapshotShowOutput{
					Name:  params.NewSnapshotName,
					Size:  "0 GiB",
					State: "created",
					Labels: map[string]string{
						"new": "label",
					},
				},
			},
			{
				Description: "update snapshot labels and name",
				Command: "exo compute block-storage snapshot update {{.NewSnapshotName}}" +
					" --label another=change" +
					" --rename {{.SnapshotName}}",
				Expected: blockStorageSnapshotShowOutput{
					Name:  params.SnapshotName,
					Size:  "0 GiB",
					State: "created",
					Labels: map[string]string{
						"another": "change",
					},
				},
			},
			{
				Description: "clear snapshot labels",
				Command: "exo compute block-storage snapshot update {{.SnapshotName}}" +
					" --label=[=]",
				Expected: blockStorageSnapshotShowOutput{
					Name:   params.SnapshotName,
					Size:   "0 GiB",
					State:  "created",
					Labels: map[string]string{},
				},
			},
			{
				Description: "clear volume labels",
				Command: "exo compute block-storage update {{.NewVolumeName}}" +
					" --label=[=]",
				Expected: blockStorageShowOutput{
					Name:   params.NewVolumeName,
					Size:   "12 GiB",
					State:  "detached",
					Labels: map[string]string{},
				},
			},
			{
				Description: "clean up snapshot",
				Command: "exo compute block-storage snapshot delete {{.SnapshotName}}" +
					" --force",
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
