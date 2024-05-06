package integ_test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO make test binary configurable

func TestMain(m *testing.M) {
	m.Run()
}

type TestCase struct {
	Command       string
	ExpectedJSON  string
	ExpectedError *exec.ExitError
}

func TestBlockStorageCreate(t *testing.T) {
	testBinary := "../../bin/exo"
	cases := []TestCase{
		{
			Command:      "exo -z ch-gva-2 -O json c bs list",
			ExpectedJSON: `{}`,
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

		assert.JSONEq(t, tc.ExpectedJSON, string(output))
	}
}
