package egoscale

import (
	"fmt"
)

// ListRequest builds the ListTemplates request
func (temp *Template) ListRequest() (ListCommand, error) {
	req := &ListTemplates{
		Name:       temp.Name,
		Account:    temp.Account,
		DomainID:   temp.DomainID,
		ID:         temp.ID,
		ZoneID:     temp.ZoneID,
		Hypervisor: temp.Hypervisor,
		//TODO Tags
	}
	if temp.IsFeatured {
		req.TemplateFilter = "featured"
	}
	if temp.Removed != "" {
		*req.ShowRemoved = true
	}

	return req, nil
}

func (*ListTemplates) each(resp interface{}, callback IterateItemFunc) {
	temps, ok := resp.(*ListTemplatesResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type. ListTemplatesResponse expected, got %T", resp))
		return
	}

	for i := range temps.Template {
		if !callback(&temps.Template[i], nil) {
			break
		}
	}
}

// SetPage sets the current page
func (ls *ListTemplates) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListTemplates) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// ResourceType returns the type of the resource
func (*Template) ResourceType() string {
	return "Template"
}

func (*ListTemplates) name() string {
	return "listTemplates"
}

func (*ListTemplates) description() string {
	return "List all public, private, and privileged templates."
}

func (*ListTemplates) response() interface{} {
	return new(ListTemplatesResponse)
}

func (*CreateTemplate) name() string {
	return "createTemplate"
}

func (*CreateTemplate) description() string {
	return "Creates a template of a virtual machine. The virtual machine must be in a STOPPED state. A template created from this command is automatically designated as a private template visible to the account that created it."
}

func (*CreateTemplate) asyncResponse() interface{} {
	return new(Template)
}

func (*PrepareTemplate) name() string {
	return "prepareTemplate"
}

func (*PrepareTemplate) description() string {
	return "load template into primary storage"
}

func (*PrepareTemplate) asyncResponse() interface{} {
	return new(Template)
}

func (*CopyTemplate) name() string {
	return "copyTemplate"
}

func (*CopyTemplate) description() string {
	return "Copies a template from one zone to another."
}

func (*CopyTemplate) asyncResponse() interface{} {
	return new(Template)
}

func (*UpdateTemplate) name() string {
	return "updateTemplate"
}

func (*UpdateTemplate) description() string {
	return "Updates attributes of a template."
}

func (*UpdateTemplate) asyncResponse() interface{} {
	return new(Template)
}

func (*DeleteTemplate) name() string {
	return "deleteTemplate"
}

func (*DeleteTemplate) description() string {
	return "Deletes a template from the system. All virtual machines using the deleted template will not be affected."
}

func (*DeleteTemplate) asyncResponse() interface{} {
	return new(booleanResponse)
}

func (*RegisterTemplate) name() string {
	return "registerTemplate"
}

func (*RegisterTemplate) description() string {
	return "Registers an existing template into the CloudStack cloud."
}

func (*RegisterTemplate) response() interface{} {
	return new(Template)
}
