package integ

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"text/template"

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
	Parameters any
	Zone       string
	Steps      []Step
	T          *testing.T
}

func (s *Suite) Run() {
	nSteps := len(s.Steps)

	for nr, step := range s.Steps {
		tmpl := template.New(fmt.Sprintf("step-%d", nr+1))
		parsedTmpl, err := tmpl.Parse(step.Command)
		if err != nil {
			s.T.Errorf("failed to parse template: %s", err.Error())

			return
		}

		buf := &bytes.Buffer{}
		err = parsedTmpl.Execute(buf, s.Parameters)
		if err != nil {
			s.T.Errorf("failed to execute template: %s", err.Error())

			return
		}

		newCommand := buf.String()
		if strings.Contains(newCommand, "{{") {
			s.T.Errorf("template execution didn't replace all parameters for step %d", nr+1)

			return
		}

		s.Steps[nr].Command = newCommand
	}

	for nr, step := range s.Steps {
		errMsg := fmt.Sprintf("step %d/%d: %s\n", nr+1, nSteps, step.Description)

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
