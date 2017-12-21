package egoscale

// ResourceTag is a tag associated with a resource
type ResourceTag struct {
	Account      string `json:"account,omitempty"`
	Customer     string `json:"customer,omitempty"`
	Domain       string `json:"domain,omitempty"`
	DomainId     string `json:"domainid,omitempty"`
	Key          string `json:"key,omitempty"`
	Project      string `json:"project,omitempty"`
	ProjectId    string `json:"projectid,omitempty"`
	ResourceId   string `json:"resourceid,omitempty"`
	ResourceType string `json:"resourcetype,omitempty"`
	Value        string `json:"value,omitempty"`
}
