package utils

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/exoscale/cli/pkg/globalstate"
)

// A failed mock call ends the goroutine with runtime.Goexit: the test must then
// fail, not hang.
func TestDecorateAsyncDoesNotHangOnGoexit(t *testing.T) {
	oldQuiet := globalstate.Quiet
	t.Cleanup(func() { globalstate.Quiet = oldQuiet })
	globalstate.Quiet = true

	tests := map[string]func(){
		"DecorateAsyncOperation": func() {
			DecorateAsyncOperation("working", func() { runtime.Goexit() })
		},
		"DecorateAsyncOperations": func() {
			_ = DecorateAsyncOperations("working", func() error {
				runtime.Goexit()
				return nil
			})
		},
	}

	for name, run := range tests {
		t.Run(name, func(t *testing.T) {
			done := make(chan struct{})
			go func() {
				defer close(done)
				run()
			}()

			select {
			case <-done:
			case <-time.After(10 * time.Second):
				assert.Fail(t, "still waiting for the goroutine")
			}
		})
	}
}
