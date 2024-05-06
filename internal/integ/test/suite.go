package test

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	Binary = "../../bin/exo"
)

type Step struct {
	Description   string
	Command       string
	Expected      interface{}
	ExpectedError *exec.ExitError
}

type Suite struct {
	Zone  string
	Steps []Step
	T     *testing.T
}

func (s *Suite) Run() {
	nSteps := len(s.Steps)
	for nr, step := range s.Steps {
		nr := nr + 1
		errMsg := fmt.Sprintf("step %d/%d: %s\n", nr, nSteps, step.Description)

		commandParts := strings.Split(step.Command, " ")
		commandParts = append([]string{"-z", s.Zone, "-O", "json"}, commandParts[1:]...)
		cmd := exec.Command(Binary, commandParts...)
		output, err := cmd.CombinedOutput()
		if step.ExpectedError != nil && err != nil {
			assert.Equal(s.T, step.ExpectedError, err, errMsg)
		} else if err != nil {
			assert.NoError(s.T, err, "unexpected error: "+string(output), errMsg)
		}

		if step.Expected == nil {
			continue
		}

		expectedType := reflect.TypeOf(step.Expected)
		actualValue := reflect.New(expectedType)
		actual := actualValue.Interface()
		err = json.Unmarshal(output, actual)
		assert.NoError(s.T, err, errMsg)

		expectedValue := reflect.ValueOf(step.Expected)
		expectedPointer := reflect.New(expectedType)
		expectedPointer.Elem().Set(expectedValue)
		expected := expectedPointer.Interface()

		assert.EqualValues(s.T, expected, actual, errMsg)
	}

}
