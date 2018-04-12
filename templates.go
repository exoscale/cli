package egoscale

// ListRequest builds the ListTemplates request
func (temp *Template) ListRequest() (ListCommand, error) {
	req := &ListTemplates{
		Name:       temp.Name,
		Account:    temp.Account,
		DomainID:   temp.DomainID,
		ID:         temp.ID,
		ProjectID:  temp.ProjectID,
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
	temps := resp.(*ListTemplatesResponse)
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

// APIName returns the CloudStack API command name
func (*ListTemplates) APIName() string {
	return "listTemplates"
}

func (*ListTemplates) response() interface{} {
	return new(ListTemplatesResponse)
}
