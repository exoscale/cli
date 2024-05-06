package test

import (
	"encoding/json"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Suite struct {
	Zone  string
	Cases []Case
	T     *testing.T
}

func (s *Suite) Run() {
	for _, testCase := range s.Cases {
		commandParts := strings.Split(testCase.Command, " ")
		commandParts = append([]string{"-z", s.Zone, "-O", "json"}, commandParts[1:]...)
		cmd := exec.Command(Binary, commandParts...)
		output, err := cmd.CombinedOutput()
		if testCase.ExpectedError != nil && err != nil {
			assert.Equal(s.T, testCase.ExpectedError, err)
		} else if err != nil {
			assert.NoError(s.T, err, "unexpected error: "+string(output))
		}

		if testCase.Expected == nil {
			continue
		}

		expectedType := reflect.TypeOf(testCase.Expected)
		actualValue := reflect.New(expectedType)
		actual := actualValue.Interface()
		err = json.Unmarshal(output, actual)
		assert.NoError(s.T, err)

		expectedValue := reflect.ValueOf(testCase.Expected)
		expectedPointer := reflect.New(expectedType)
		expectedPointer.Elem().Set(expectedValue)
		expected := expectedPointer.Interface()

		assert.EqualValues(s.T, expected, actual)
	}

}
