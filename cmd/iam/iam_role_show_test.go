package iam

import (
	"net/http"
	"net/http/httptest"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
)

func TestIAMRoleShowIncludesPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/iam-role" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		testutils.WriteJSON(t, w, http.StatusOK, v3.ListIAMRolesResponse{IAMRoles: []v3.IAMRole{{
			ID:   v3.UUID("0e324aaf-4b7e-45ff-a8c9-5830e19484cc"),
			Name: "vpc-canary-iam-role",
			Policy: &v3.IAMPolicy{
				DefaultServiceStrategy: v3.IAMPolicyDefaultServiceStrategyDeny,
				Services: map[string]v3.IAMServicePolicy{
					"networking": {Type: v3.IAMServicePolicyTypeAllow},
					"compute":    {Type: v3.IAMServicePolicyTypeAllow},
				},
			},
		}}})
	}))
	defer server.Close()

	testutils.SetupV3Client(t, server.URL)
	account.CurrentAccount = &account.Account{}
	cmd := &iamRoleShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Role:               "vpc-canary-iam-role",
	}
	var got *iamRoleShowOutput
	cmd.OutputFunc = func(o output.Outputter, err error) error {
		got = o.(*iamRoleShowOutput)
		return err
	}

	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatal(err)
	}
	if got.Policy == nil || got.Policy.DefaultServiceStrategy != "deny" {
		t.Fatalf("unexpected policy: %+v", got.Policy)
	}
	if got.Policy.Services["networking"].Type != "allow" || got.Policy.Services["compute"].Type != "allow" {
		t.Fatalf("unexpected policy services: %+v", got.Policy.Services)
	}
}
