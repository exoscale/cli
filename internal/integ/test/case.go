package test

import "os/exec"

var (
	Binary = "../../bin/exo"
)

type Case struct {
	Command       string
	Expected      interface{}
	ExpectedError *exec.ExitError
}
