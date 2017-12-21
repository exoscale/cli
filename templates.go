/*
Templates

See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/latest/templates.html
*/

package egoscale

// Template represents a machine to be deployed
type Template struct {
	Account               string            `json:"account,omitempty"`
	AccountId             string            `json:"accountid,omitempty"`
	Bootable              bool              `json:"bootable,omitempty"`
	Checksum              string            `json:"checksum,omitempty"`
	CreatedAt             string            `json:"created,omitempty"`
	CrossZones            bool              `json:"crossZones,omitempty"`
	Details               map[string]string `json:"details,omitempty"`
	DisplayText           string            `json:"displaytext,omitempty"`
	Domain                string            `json:"domain,omitempty"`
	DomainId              string            `json:"domainid,omitempty"`
	Format                string            `json:"format,omitempty"`
	HostId                string            `json:"hostid,omitempty"`
	HostName              string            `json:"hostname,omitempty"`
	Hypervisor            string            `json:"hypervisor,omitempty"`
	Id                    string            `json:"id,omitempty"`
	IsDynamicallyScalable bool              `json:"isdynamicallyscalable,omitempty"`
	IsExtractable         bool              `json:"isextractable,omitempty"`
	IsFeatured            bool              `json:"isfeatured,omitempty"`
	IsPublic              bool              `json:"ispublic,omitempty"`
	IsReady               bool              `json:"isready,omitempty"`
	Name                  string            `json:"name,omitempty"`
	OsTypeId              string            `json:"ostypeid,omitempty"`
	OsTypeName            string            `json:"ostypename,omitempty"`
	PasswordEnabled       bool              `json:"passwordenabled,omitempty"`
	Project               string            `json:"project,omitempty"`
	ProjectId             string            `json:"projectid,omitempty"`
	RemovedAt             string            `json:"removed,omitempty"`
	Size                  int64             `json:"size,omitempty"`
	SourceTemplateId      string            `json:"sourcetemplateid,omitempty"`
	SshKeyEnabled         bool              `json:"sshkeyenabled,omitempty"`
	Status                string            `json:"status,omitempty"`
	Zoneid                string            `json:"zoneid,omitempty"`
	Zonename              string            `json:"zonename,omitempty"`
}

// ListTemplatesRequest represents a template query filter
type ListTemplatesRequest struct {
	TemplateFilter string         `json:"templatefilter"` // featured, etc.
	Account        string         `json:"account,omitempty"`
	DomainId       string         `json:"domainid,omitempty"`
	Hypervisor     string         `json:"hypervisor,omitempty"`
	Id             string         `json:"id,omitempty"`
	IsRecursive    bool           `json:"isrecursive,omitempty"`
	Keyword        string         `json:"keyword,omitempty"`
	ListAll        bool           `json:"listall,omitempty"`
	Name           string         `json:"name,omitempty"`
	Page           int            `json:"page,omitempty"`
	PageSize       int            `json:"pagesize,omitempty"`
	ProjectId      string         `json:"projectid,omitempty"`
	ShowRemoved    bool           `json:"showremoved,omitempty"`
	Tags           []*ResourceTag `json:"tags,omitempty"`
	ZoneId         string         `json:"zoneid,omitempty"`
}

// Command returns the CloudStack API command
func (req *ListTemplatesRequest) Command() string {
	return "listTemplates"
}

// ListTemplatesResponse represents a list of templates
type ListTemplatesResponse struct {
	Count    int         `json:"count"`
	Template []*Template `json:"template"`
}
