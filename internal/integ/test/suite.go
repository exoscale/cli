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
	Cases []Case
	T     *testing.T
}

func (s *Suite) Run() {
	for _, tc := range s.Cases {
		commandParts := strings.Split(tc.Command, " ")
		cmd := exec.Command(Binary, commandParts[1:]...)
		output, err := cmd.CombinedOutput()
		if tc.ExpectedError != nil && err != nil {
			assert.Equal(s.T, tc.ExpectedError, err)
		} else if err != nil {
			assert.NoError(s.T, err, "unexpected error: "+string(output))
		}

		expectedType := reflect.TypeOf(tc.Expected)
		actualValue := reflect.New(expectedType)
		actual := actualValue.Interface()
		err = json.Unmarshal(output, actual)
		assert.NoError(s.T, err)

		expectedValue := reflect.ValueOf(tc.Expected)
		expectedPointer := reflect.New(expectedType)
		expectedPointer.Elem().Set(expectedValue)
		expected := expectedPointer.Interface()

		assert.EqualValues(s.T, expected, actual)
	}

}
