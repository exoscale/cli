//go:build integration_api
// +build integration_api

package integration_with_api_test

// TestStoragePurgeObjectLock reproduces "exo storage purge infinite loop on bucket with object lock enabled".
// Requires EXOSCALE_API_KEY, EXOSCALE_API_SECRET, EXOSCALE_DEFAULT_ZONE, aws CLI in PATH, and bin/exo (make build).
// Run: go test -v -tags=integration_api -run TestStoragePurgeObjectLock ./tests/integ/with-api/

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	purgeTTL      = 10 * time.Second
	retentionDays = 1
)

func TestStoragePurgeObjectLock(t *testing.T) {
	apiKey := requireEnv(t, "EXOSCALE_API_KEY")
	apiSecret := requireEnv(t, "EXOSCALE_API_SECRET")
	zone := requireEnv(t, "EXOSCALE_DEFAULT_ZONE")

	sosEndpoint := os.Getenv("EXOSCALE_SOS_ENDPOINT")
	if sosEndpoint == "" {
		sosEndpoint = fmt.Sprintf("https://sos-%s.exo.io", zone)
	}

	bucket := fmt.Sprintf("integ-purge-lock-%d", rand.Int63n(1_000_000))

	awsEnv := append(
		os.Environ(),
		"AWS_ACCESS_KEY_ID="+apiKey,
		"AWS_SECRET_ACCESS_KEY="+apiSecret,
		"AWS_DEFAULT_REGION="+zone,
	)

	runAWS := func(t *testing.T, args ...string) []byte {
		t.Helper()
		fullArgs := append([]string{"s3api", "--endpoint-url", sosEndpoint}, args...)
		cmd := exec.Command("aws", fullArgs...)
		cmd.Env = awsEnv
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "aws s3api %s failed:\n%s", strings.Join(args[:1], " "), out)
		return out
	}

	t.Logf("step 1: create bucket %q with object lock enabled", bucket)
	runAWS(t,
		"create-bucket",
		"--bucket", bucket,
		"--object-lock-enabled-for-bucket",
	)

	t.Cleanup(func() { cleanupLockedBucket(t, bucket, sosEndpoint, awsEnv) })

	t.Log("step 2: apply GOVERNANCE retention policy (1 day)")
	lockCfg, _ := json.Marshal(map[string]any{
		"ObjectLockEnabled": "Enabled",
		"Rule": map[string]any{
			"DefaultRetention": map[string]any{
				"Mode": "GOVERNANCE",
				"Days": retentionDays,
			},
		},
	})
	runAWS(t,
		"put-object-lock-configuration",
		"--bucket", bucket,
		"--object-lock-configuration", string(lockCfg),
	)

	t.Log("step 3: upload test objects via exo CLI")
	f1 := writeTempFile(t, "hello-*.txt", "hello world\n")
	f2 := writeTempFile(t, "doom_rip_off-*.jpeg", "not a real jpeg, just test data\n")

	exoBin := "../../../bin/exo"
	uploadCmd := exec.Command(exoBin,
		"-z", zone,
		"storage", "upload",
		f1, f2,
		fmt.Sprintf("sos://%s/test/", bucket),
	)
	uploadOut, err := uploadCmd.CombinedOutput()
	require.NoError(t, err, "exo storage upload failed:\n%s", uploadOut)
	t.Logf("uploaded:\n%s", uploadOut)

	t.Logf("step 4: exo storage purge sos://%s/ (timeout=%s)", bucket, purgeTTL)

	ctx, cancel := context.WithTimeout(context.Background(), purgeTTL)
	defer cancel()

	purgeCmd := exec.CommandContext(ctx,
		exoBin,
		"-z", zone,
		"storage", "purge",
		"--force",
		fmt.Sprintf("sos://%s/", bucket),
	)

	var stdout, stderr bytes.Buffer
	purgeCmd.Stdout = &stdout
	purgeCmd.Stderr = &stderr

	_ = purgeCmd.Run()

	purgeOutput := stdout.String() + stderr.String()
	t.Logf("purge output:\n%s", purgeOutput)

	// If the deadline fired the process never exited on its own.
	require.NoError(t, ctx.Err(),
		"exo storage purge did not terminate within %s.\n"+
			"It is stuck in an infinite loop re-listing objects that cannot be\n"+
			"deleted because of object lock.\n\nOutput:\n%s",
		purgeTTL, purgeOutput,
	)

	// Errors should be human-readable, not raw pointer addresses like:
	//   Error happened: delete error: {<nil> 0xc332882c390 0xc332882c3b0 0xc332882c3a0}
	require.False(t, containsPointerAddress(purgeOutput),
		"exo storage purge printed raw pointer addresses instead of human-readable\n"+
			"error messages; the *string fields in types.Error are not being dereferenced.\n\n"+
			"Output:\n%s",
		purgeOutput,
	)
}

func requireEnv(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Skipf("skipping: %s env var not set", key)
	}
	return v
}

func writeTempFile(t *testing.T, pattern, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", pattern)
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(f.Name()) })
	_, err = fmt.Fprint(f, content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

func containsPointerAddress(s string) bool {
	for i := 0; i+2 < len(s); i++ {
		if s[i] != '0' || s[i+1] != 'x' {
			continue
		}
		hexLen := 0
		for j := i + 2; j < len(s); j++ {
			c := s[j]
			if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
				hexLen++
			} else {
				break
			}
		}
		if hexLen >= 8 {
			return true
		}
	}
	return false
}

func cleanupLockedBucket(t *testing.T, bucket, sosEndpoint string, awsEnv []string) {
	t.Helper()
	t.Logf("cleanup: removing bucket %s", bucket)

	run := func(args ...string) ([]byte, error) {
		fullArgs := append([]string{"s3api", "--endpoint-url", sosEndpoint}, args...)
		cmd := exec.Command("aws", fullArgs...)
		cmd.Env = awsEnv
		return cmd.CombinedOutput()
	}

	out, err := run("list-object-versions", "--bucket", bucket)
	if err != nil {
		t.Logf("cleanup: list-object-versions failed: %v\n%s", err, out)
		return
	}

	var listing struct {
		Versions []struct {
			Key       string `json:"Key"`
			VersionID string `json:"VersionId"`
		} `json:"Versions"`
		DeleteMarkers []struct {
			Key       string `json:"Key"`
			VersionID string `json:"VersionId"`
		} `json:"DeleteMarkers"`
	}
	if err := json.Unmarshal(out, &listing); err != nil {
		t.Logf("cleanup: could not parse version listing: %v", err)
		return
	}

	type objectID struct {
		Key       string `json:"Key"`
		VersionID string `json:"VersionId"`
	}
	var objects []objectID
	for _, v := range listing.Versions {
		objects = append(objects, objectID{Key: v.Key, VersionID: v.VersionID})
	}
	for _, d := range listing.DeleteMarkers {
		objects = append(objects, objectID{Key: d.Key, VersionID: d.VersionID})
	}

	if len(objects) > 0 {
		deletePayload, _ := json.Marshal(map[string]any{
			"Objects": objects,
			"Quiet":   false,
		})
		out, err = run(
			"delete-objects",
			"--bucket", bucket,
			"--delete", string(deletePayload),
			"--bypass-governance-retention",
		)
		if err != nil {
			t.Logf("cleanup: delete-objects failed: %v\n%s", err, out)
			return
		}
	}

	out, err = run("delete-bucket", "--bucket", bucket)
	if err != nil {
		t.Logf("cleanup: delete-bucket failed (will expire naturally in %d day(s)): %v\n%s",
			retentionDays, err, out)
	}
}
