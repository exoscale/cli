package compute

import (
	"context"
	"fmt"

	v3 "github.com/exoscale/egoscale/v3"
)

// ResolveTemplateID passes UUIDs through and resolves template names from the list endpoint.
func ResolveTemplateID(
	ctx context.Context,
	client *v3.Client,
	nameOrID string,
	visibility string,
	zone v3.ZoneName,
) (v3.UUID, error) {
	if id, err := v3.ParseUUID(nameOrID); err == nil {
		return id, nil
	}

	templates, err := client.ListTemplates(ctx, v3.ListTemplatesWithVisibility(v3.ListTemplatesVisibility(visibility)))
	if err != nil {
		return "", fmt.Errorf("error listing template with visibility %q: %w", visibility, err)
	}
	template, err := templates.FindTemplate(nameOrID)
	if err != nil {
		return "", fmt.Errorf(
			"no template %q found with visibility %s in zone %s",
			nameOrID,
			visibility,
			zone,
		)
	}

	return template.ID, nil
}
