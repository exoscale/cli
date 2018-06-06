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

func (*ListTemplates) response() interface{} {
	return new(ListTemplatesResponse)
}

func (*CreateTemplate) name() string {
	return "createTemplate"
}

func (*CreateTemplate) asyncResponse() interface{} {
	return new(Template)
}

func (*PrepareTemplate) name() string {
	return "prepareTemplate"
}

func (*PrepareTemplate) asyncResponse() interface{} {
	return new(Template)
}

func (*CopyTemplate) name() string {
	return "copyTemplate"
}

func (*CopyTemplate) asyncResponse() interface{} {
	return new(Template)
}

func (*UpdateTemplate) name() string {
	return "updateTemplate"
}

func (*UpdateTemplate) asyncResponse() interface{} {
	return new(Template)
}

func (*DeleteTemplate) name() string {
	return "deleteTemplate"
}

func (*DeleteTemplate) asyncResponse() interface{} {
	return new(booleanResponse)
}

func (*RegisterTemplate) name() string {
	return "registerTemplate"
}

func (*RegisterTemplate) response() interface{} {
	return new(Template)
}
