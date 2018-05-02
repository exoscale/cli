package egoscale

// Template represents a machine to be deployed
//
// See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/latest/templates.html
type Template struct {
	Account               string            `json:"account,omitempty" doc:"the account name to which the template belongs"`
	AccountID             string            `json:"accountid,omitempty" doc:"the account id to which the template belongs"`
	Bootable              bool              `json:"bootable,omitempty" doc:"true if the ISO is bootable, false otherwise"`
	Checksum              string            `json:"checksum,omitempty" doc:"checksum of the template"`
	Created               string            `json:"created,omitempty" doc:"the date this template was created"`
	CrossZones            bool              `json:"crossZones,omitempty" doc:"true if the template is managed across all Zones, false otherwise"`
	Details               map[string]string `json:"details,omitempty" doc:"additional key/value details tied with template"`
	DisplayText           string            `json:"displaytext,omitempty" doc:"the template display text"`
	Domain                string            `json:"domain,omitempty" doc:"the name of the domain to which the template belongs"`
	DomainID              string            `json:"domainid,omitempty" doc:"the ID of the domain to which the template belongs"`
	Format                string            `json:"format,omitempty" doc:"the format of the template."`
	HostID                string            `json:"hostid,omitempty" doc:"the ID of the secondary storage host for the template"`
	HostName              string            `json:"hostname,omitempty" doc:"the name of the secondary storage host for the template"`
	Hypervisor            string            `json:"hypervisor,omitempty" doc:"the hypervisor on which the template runs"`
	ID                    string            `json:"id,omitempty" doc:"the template ID"`
	IsDynamicallyScalable bool              `json:"isdynamicallyscalable,omitempty" doc:"true if template contains XS/VMWare tools inorder to support dynamic scaling of VM cpu/memory"`
	IsExtractable         bool              `json:"isextractable,omitempty" doc:"true if the template is extractable, false otherwise"`
	IsFeatured            bool              `json:"isfeatured,omitempty" doc:"true if this template is a featured template, false otherwise"`
	IsPublic              bool              `json:"ispublic,omitempty" doc:"true if this template is a public template, false otherwise"`
	IsReady               bool              `json:"isready,omitempty" doc:"true if the template is ready to be deployed from, false otherwise."`
	Name                  string            `json:"name,omitempty" doc:"the template name"`
	OsTypeID              string            `json:"ostypeid,omitempty" doc:"the ID of the OS type for this template."`
	OsTypeName            string            `json:"ostypename,omitempty" doc:"the name of the OS type for this template."`
	PasswordEnabled       bool              `json:"passwordenabled,omitempty" doc:"true if the reset password feature is enabled, false otherwise"`
	Project               string            `json:"project,omitempty" doc:"the project name of the template"`
	ProjectID             string            `json:"projectid,omitempty" doc:"the project id of the template"`
	Removed               string            `json:"removed,omitempty" doc:"the date this template was removed"`
	Size                  int64             `json:"size,omitempty" doc:"the size of the template"`
	SourceTemplateID      string            `json:"sourcetemplateid,omitempty" doc:"the template ID of the parent template if present"`
	SSHKeyEnabled         bool              `json:"sshkeyenabled,omitempty" doc:"true if template is sshkey enabled, false otherwise"`
	Status                string            `json:"status,omitempty" doc:"the status of the template"`
	Tags                  []ResourceTag     `json:"tags,omitempty" doc:"the list of resource tags associated with tempate"`
	TemplateDirectory     string            `json:"templatedirectory,omitempty" doc:"Template directory"`
	TemplateTag           string            `json:"templatetag,omitempty" doc:"the tag of this template"`
	TemplateType          string            `json:"templatetype,omitempty" doc:"the type of the template"`
	Url                   string            `json:"url,omitempty" doc:"Original URL of the template where it was downloaded"`
	ZoneID                string            `json:"zoneid,omitempty" doc:"the ID of the zone for this template"`
	ZoneName              string            `json:"zonename,omitempty" doc:"the name of the zone for this template"`
}

// ListTemplates represents a template query filter
type ListTemplates struct {
	TemplateFilter string        `json:"templatefilter" doc:"possible values are \"featured\", \"self\", \"selfexecutable\",\"sharedexecutable\",\"executable\", and \"community\". * featured : templates that have been marked as featured and public. * self : templates that have been registered or created by the calling user. * selfexecutable : same as self, but only returns templates that can be used to deploy a new VM. * sharedexecutable : templates ready to be deployed that have been granted to the calling user by another user. * executable : templates that are owned by the calling user, or public templates, that can be used to deploy a VM. * community : templates that have been marked as public but not featured. * all : all templates (only usable by admins)."`
	Account        string        `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	DomainID       string        `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	Hypervisor     string        `json:"hypervisor,omitempty" doc:"the hypervisor for which to restrict the search"`
	ID             string        `json:"id,omitempty" doc:"the template ID"`
	IsRecursive    *bool         `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword        string        `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll        *bool         `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Name           string        `json:"name,omitempty" doc:"the template name"`
	Page           int           `json:"page,omitempty"`
	PageSize       int           `json:"pagesize,omitempty"`
	ProjectID      string        `json:"projectid,omitempty" doc:"list objects by project"`
	ShowRemoved    *bool         `json:"showremoved,omitempty" doc:"show removed templates as well"`
	Tags           []ResourceTag `json:"tags,omitempty" doc:"List resources by tags (key/value pairs)"`
	ZoneID         string        `json:"zoneid,omitempty" doc:"list templates by zoneId"`
}

// ListTemplatesResponse represents a list of templates
type ListTemplatesResponse struct {
	Count    int        `json:"count"`
	Template []Template `json:"template"`
}
