package e2e_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func TestAPIRequestInterrupt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Interrupt cannot be sent on Windows")
	}

	requestStarted := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-r.Context().Done()
	}))
	defer server.Close()

	home := t.TempDir()
	cmd := exec.Command(exoBinary, "api", "list-zones", "--server", server.URL)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	cmd.Env = []string{
		"EXOSCALE_API_KEY=test-key",
		"EXOSCALE_API_SECRET=test-secret",
		"HOME=" + home,
		"XDG_CONFIG_HOME=" + home,
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	defer func() {
		_ = cmd.Process.Kill()
		<-done
	}()

	select {
	case <-requestStarted:
	case <-done:
		t.Fatalf("exo exited before making the API request:\n%s", output.String())
	case <-time.After(10 * time.Second):
		t.Fatal("API request did not reach the server")
	}

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("exo did not exit after SIGINT")
	}
}
