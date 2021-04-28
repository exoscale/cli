package v2

import (
	"context"
	"time"

	apiv2 "github.com/exoscale/egoscale/v2/api"
	papi "github.com/exoscale/egoscale/v2/internal/public-api"
)

// Template represents a Compute instance template.
type Template struct {
	BootMode        string
	Build           string
	Checksum        string
	CreatedAt       time.Time
	DefaultUser     string
	Description     string
	Family          string
	ID              string
	Name            string
	PasswordEnabled bool
	SSHKeyEnabled   bool
	Size            int64
	URL             string
	Version         string
	Visibility      string
}

func templateFromAPI(t *papi.Template) *Template {
	return &Template{
		BootMode:        papi.OptionalString(t.BootMode),
		Build:           papi.OptionalString(t.Build),
		Checksum:        papi.OptionalString(t.Checksum),
		CreatedAt:       *t.CreatedAt,
		DefaultUser:     papi.OptionalString(t.DefaultUser),
		Description:     papi.OptionalString(t.Description),
		Family:          papi.OptionalString(t.Family),
		ID:              papi.OptionalString(t.Id),
		Name:            papi.OptionalString(t.Name),
		PasswordEnabled: papi.OptionalBool(t.PasswordEnabled),
		SSHKeyEnabled:   papi.OptionalBool(t.SshKeyEnabled),
		Size:            *t.Size,
		URL:             papi.OptionalString(t.Url),
		Version:         papi.OptionalString(t.Version),
		Visibility:      papi.OptionalString(t.Visibility),
	}
}

// ListTemplates returns the list of existing Templates in the specified zone.
func (c *Client) ListTemplates(ctx context.Context, zone, visibility, family string) ([]*Template, error) {
	list := make([]*Template, 0)

	resp, err := c.ListTemplatesWithResponse(apiv2.WithZone(ctx, zone), &papi.ListTemplatesParams{
		Visibility: &visibility,
		Family: func() *string {
			if family != "" {
				return &family
			}
			return nil
		}(),
	})
	if err != nil {
		return nil, err
	}

	if resp.JSON200.Templates != nil {
		for i := range *resp.JSON200.Templates {
			list = append(list, templateFromAPI(&(*resp.JSON200.Templates)[i]))
		}
	}

	return list, nil
}

// GetTemplate returns the Template corresponding to the specified ID in the specified zone.
func (c *Client) GetTemplate(ctx context.Context, zone, id string) (*Template, error) {
	resp, err := c.GetTemplateWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}

	return templateFromAPI(resp.JSON200), nil
}

// DeleteTemplate deletes the specified Template in the specified zone.
func (c *Client) DeleteTemplate(ctx context.Context, zone, id string) error {
	resp, err := c.DeleteTemplateWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.timeout).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}
