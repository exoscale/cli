package integ_test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

type TestCase struct {
	Command       string
	Expected      any
	ExpectedError *exec.ExitError
}

func TestBlockStorageCreate(t *testing.T) {
	testBinary := "../../bin/exo"
	cases := []TestCase{
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

	for _, tc := range cases {
		commandParts := strings.Split(tc.Command, " ")
		cmd := exec.Command(testBinary, commandParts[1:]...)
		output, err := cmd.CombinedOutput()
		if tc.ExpectedError != nil && err != nil {
			assert.Equal(t, tc.ExpectedError, err)
		} else if err != nil {
			assert.NoError(t, err, "unexpected error: "+string(output))
		}

		out := make([]blockStorageListItemOutput, 0)
		err = json.Unmarshal(output, &out)
		assert.NoError(t, err)

		assert.Equal(t, tc.Expected, out)
	}
}
