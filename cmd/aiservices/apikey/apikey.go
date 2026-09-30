package apikey

import (
	"fmt"

	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

// Cmd is the root command for api-key subcommands.
var Cmd = &cobra.Command{
	Use:   "api-key",
	Short: "Manage AI API keys",
}

func conflictingFlagsErr(a, b string) error {
	return fmt.Errorf("--%s cannot be used together with --%s", a, b)
}

func derefBool(b *bool) bool {
	return b != nil && *b
}

func modelsToStrings(m *v3.AIAPIKeyModels) []string {
	if m == nil {
		return nil
	}
	out := make([]string, len(*m))
	copy(out, *m)
	return out
}

func deploymentsToStrings(d *v3.AIAPIKeyDeployments) []string {
	if d == nil {
		return nil
	}
	out := make([]string, 0, len(*d))
	for _, item := range *d {
		out = append(out, item.ID.String())
	}
	return out
}

func deploymentsToRefs(ids []string) (v3.AIAPIKeyDeployments, error) {
	refs := make(v3.AIAPIKeyDeployments, 0, len(ids))
	for _, id := range ids {
		uuid, err := v3.ParseUUID(id)
		if err != nil {
			return nil, fmt.Errorf("invalid deployment ID %q: must be a UUID", id)
		}
		refs = append(refs, v3.AIAPIKeyDeploymentRef{ID: uuid})
	}
	return refs, nil
}
