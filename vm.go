/*
Virtual Machines

... todo ...

See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/stable/virtual_machines.html
*/
package egoscale

// VirtualMachine reprents a virtual machine
type VirtualMachine struct {
	Id                    string            `json:"id,omitempty"`
	Account               string            `json:"account,omitempty"`
	ClusterId             string            `json:"clusterid,omitempty"`
	ClusterName           string            `json:"clustername,omitempty"`
	CpuNumber             int64             `json:"cpunumber,omitempty"`
	CpuSpeed              int64             `json:"cpuspeed,omitempty"`
	CpuUsed               string            `json:"cpuused,omitempty"`
	Created               string            `json:"created,omitempty"`
	Details               map[string]string `json:"details,omitempty"`
	DiskIoRead            int64             `json:"diskioread,omitempty"`
	DiskIoWrite           int64             `json:"diskiowrite,omitempty"`
	DiskKbsRead           int64             `json:"diskkbsread,omitempty"`
	DiskKbsWrite          int64             `json:"diskkbswrite,omitempty"`
	DiskOfferingId        string            `json:"diskofferingid,omitempty"`
	DiskOfferingName      string            `json:"diskofferingname,omitempty"`
	DisplayName           string            `json:"displayname,omitempty"`
	DisplayVm             bool              `json:"displayvm,omitempty"`
	Domain                string            `json:"domain,omitempty"`
	DomainId              string            `json:"domainid,omitempty"`
	ForVirtualNetwork     bool              `json:"forvirtualnetwork,omitempty"`
	Group                 string            `json:"group,omitempty"`
	GroupId               string            `json:"groupid,omitempty"`
	GuestOsId             string            `json:"guestosid,omitempty"`
	HaEnable              bool              `json:"haenable,omitempty"`
	HostId                string            `json:"hostid,omitempty"`
	HostName              string            `json:"hostname,omitempty"`
	Hypervisor            string            `json:"hypervisor,omitempty"`
	InstanceName          string            `json:"instancename,omitempty"` // root only
	IsDynamicallyScalable bool              `json:"isdynamicallyscalable,omitempty"`
	IsoDisplayText        string            `json:"isodisplaytext,omitempty"`
	IsoId                 string            `json:"isoid,omitempty"`
	IsoName               string            `json:"isoname,omitempty"`
	KeyPair               string            `json:"keypair,omitempty"`
	Memory                int64             `json:"memory,omitempty"`
	MemoryIntFreeKbs      int64             `json:"memoryintfreekbs,omitempty"`
	MemoryKbs             int64             `json:"memorykbs,omitempty"`
	MemoryTargetKbs       int64             `json:"memorytargetkbs,omitempty"`
	Name                  string            `json:"name,omitempty"`
	NetworkKbsRead        int64             `json:"networkkbsread,omitempty"`
	NetworkKbsWrite       int64             `json:"networkkbswrite,omitempty"`
	OsCategoryId          string            `json:"oscategoryid,omitempty"`
	OsTypeId              string            `json:"ostypeid,omitempty"`
	Password              string            `json:"password,omitempty"`
	PasswordEnabled       bool              `json:"passwordenabled,omitempty"`
	PciDevices            string            `json:"pcidevices,omitempty"` // not in the doc
	PodId                 string            `json:"podid,omitempty"`
	PodName               string            `json:"podname,omitempty"`
	Project               string            `json:"project,omitempty"`
	ProjectId             string            `json:"projectid,omitempty"`
	PublicIp              string            `json:"publicip,omitempty"`
	PublicIpId            string            `json:"publicipid,omitempty"`
	RootDeviceTd          int64             `json:"rootdeviceid,omitempty"`
	RootDeviceType        string            `json:"rootdevicetype,omitempty"`
	ServiceOfferingId     string            `json:"serviceofferingid,omitempty"`
	ServiceOfferingName   string            `json:"serviceofferingname,omitempty"`
	ServiceState          string            `json:"servicestate,omitempty"`
	State                 string            `json:"state,omitempty"`
	TemplateDisplayText   string            `json:"templatedisplaytext,omitempty"`
	TemplateId            string            `json:"templateid,omitempty"`
	TemplateName          string            `json:"templatename,omitempty"`
	UserId                string            `json:"userid,omitempty"`   // not in the doc
	UserName              string            `json:"username,omitempty"` // not in the doc
	Vgpu                  string            `json:"vgpu,omitempty"`     // not in the doc
	ZoneId                string            `json:"zoneid,omitempty"`
	ZoneName              string            `json:"zonename,omitempty"`
	AffinityGroup         []*AffinityGroup  `json:"affinitygroup,omitempty"`
	Nic                   []*Nic            `json:"nic,omitempty"`
	SecurityGroup         []*SecurityGroup  `json:"securitygroup,omitempty"`
	Tags                  []*ResourceTag    `json:"tags,omitempty"`
	JobId                 string            `json:"jobid,omitempty"`
	JobStatus             JobStatusType     `json:"jobstatus,omitempty"`
}

// DeployVirtualMachineRequest represents the machine creation
type DeployVirtualMachineRequest struct {
	ServiceOfferingId  string         `json:"serviceofferingid"`
	TemplateId         string         `json:"templateid"`
	ZoneId             string         `json:"zoneid"`
	Account            string         `json:"account,omitempty"`
	AffinityGroupIds   string         `json:"affinitygroupids,omitempty"`   // comma separated list, mutually exclusive with names
	AffinityGroupNames string         `json:"affinitygroupnames,omitempty"` // comma separated list, mutually exclusive with ids
	CustomId           string         `json:"customid,omitempty"`           // root only
	DeploymentPlanner  string         `json:"deploymentplanner,omitempty"`  // root only
	Details            string         `json:"details,omitempty"`
	DiskOfferingId     string         `json:"diskofferingid,omitempty"`
	DisplayName        string         `json:"displayname,omitempty"`
	DisplayVM          bool           `json:"displayvm,omitempty"`
	DomainId           string         `json:"domainid,omitempty"`
	Group              string         `json:"group,omitempty"`
	HostId             string         `json:"hostid,omitempty"`
	Hypervisor         string         `json:"hypervisor,omitempty"`
	Ip6Address         string         `json:"ip6address,omitempty"`
	IpAddress          string         `json:"ipaddress,omitempty"`
	IpToNetworkList    []*IpToNetwork `json:"iptonetworklist,omitempty"`
	Keyboard           string         `json:"keyboard,omitempty"`
	KeyPair            string         `json:"keypair,omitempty"`
	Name               string         `json:"name,omitempty"`
	NetworkIds         []string       `json:"networkids,omitempty"` // mutually exclusive with iptonetworklist
	ProjectId          string         `json:"projectid,omitempty"`
	RootDiskSize       int64          `json:"rootdisksize,omitempty"`       // in GiB
	SecurityGroupIds   string         `json:"securitygroupids,omitempty"`   // comma separated list, exclusive with names
	SecurityGroupNames string         `json:"securitygroupnames,omitempty"` // comma separated list, exclusive with ids
	Size               string         `json:"size,omitempty"`               // mutually exclusive with diskofferingid
	StartVm            bool           `json:"startvm,omitempty"`
	UserData           []byte         `json:"userdata,omitempty"`
}

// Command returns the command name for the Cloud Stack API
func (req *DeployVirtualMachineRequest) Command() string {
	return "deployVirtualMachine"
}

// StartVirtualMachineRequest represents the creation of the virtual machine
type StartVirtualMachineRequest struct {
	Id               string `json:"id"`
	DeploymentPlaner string `json:"deploymentplanner,omitempty"` // root only
	HostId           string `json:"hostid,omitempty"`            // root only
}

// Command returns the command name for the Cloud Stack API
func (req *StartVirtualMachineRequest) Command() string {
	return "startVirtualMachine"
}

// StopVirtualMachineRequest represents the stopping of the virtual machine
type StopVirtualMachineRequest struct {
	Id     string `json:"id"`
	Forced bool   `json:"forced,omitempty"`
}

// Command returns the command name for the Cloud Stack API
func (req *StopVirtualMachineRequest) Command() string {
	return "stopVirtualMachine"
}

// RebootVirtualMachineRequest represents the rebooting of the virtual machine
type RebootVirtualMachineRequest struct {
	Id string `json:"id"`
}

// Command returns the command name for the Cloud Stack API
func (req *RebootVirtualMachineRequest) Command() string {
	return "rebootVirtualMachine"
}

// DestroyVirtualMachineRequest represents the destruction of the virtual machine
type DestroyVirtualMachineRequest struct {
	Id      string `json:"id"`
	Expunge bool   `json:"expunge,omitempty"`
}

// Command returns the command name for the Cloud Stack API
func (req *DestroyVirtualMachineRequest) Command() string {
	return "destroyVirtualMachine"
}

// VirtualMachineResponse represents a deployed VM instance
type VirtualMachineResponse struct {
	VirtualMachine *VirtualMachine `json:"virtualmachine"`
}

// ListVirtualMachineRequest represents a search for a VM
type ListVirtualMachinesRequest struct {
	Account           string         `json:"account,omitempty"`
	AffinityGroupId   string         `json:"affinitygroupid,omitempty"`
	Details           string         `json:"details,omitempty"`   // comma separated list, all, group, nics, stats, ...
	DisplayVm         bool           `json:"displayvm,omitempty"` // root only
	DomainId          string         `json:"domainin,omitempty"`
	ForVirtualNetwork bool           `json:"forvirtualnetwork,omitempty"`
	GroupId           string         `json:"groupid,omitempty"`
	HostId            string         `json:"hostid,omitempty"`
	Hypervisor        string         `json:"hypervisor,omitempty"`
	Id                string         `json:"id,omitempty"`
	Ids               string         `json:"ids,omitempty"` // mutually exclusive with id
	IsoId             string         `json:"isoid,omitempty"`
	IsRecursive       bool           `json:"isrecursive,omitempty"`
	KeyPair           string         `json:"keypair,omitempty"`
	Keyword           string         `json:"keyword,omitempty"`
	ListAll           bool           `json:"listall,omitempty"`
	Name              string         `json:"name,omitempty"`
	NetworkId         string         `json:"networkid,omitempty"`
	Page              int            `json:"page,omitempty"`
	PageSize          int            `json:"pagesize,omitempty"`
	PodId             string         `json:"podid,omitempty"`
	ProjectId         string         `json:"projectid,omitempty"`
	ServiceOfferindId string         `json:"serviceofferingid,omitempty"`
	State             string         `json:"state,omitempty"` // Running, Stopped, Present, ...
	StorageId         string         `json:"storageid,omitempty"`
	Tags              []*ResourceTag `json:"tags,omitempty"`
	TemplateId        string         `json:"templateid,omitempty"`
	UserId            string         `json:"userid,omitempty"`
	VpcId             string         `json:"vpcid,omitempty"`
	ZoneId            string         `json:"zoneid,omitempty"`
}

// Command returns the command name for the Cloud Stack API
func (req *ListVirtualMachinesRequest) Command() string {
	return "listVirtualMachines"
}

// ListVirtualMachinesResponse represents a list of virtual machines
type ListVirtualMachinesResponse struct {
	Count          int               `json:"count"`
	VirtualMachine []*VirtualMachine `json:"virtualmachine"`
}

// IpToNetwork represents a mapping between ip and networks
type IpToNetwork struct {
	Ip        string `json:"ip,omitempty"`
	IpV6      string `json:"ipv6,omitempty"`
	NetworkId string `json:"networkid,omitempty"`
}

// DeployVirtualMachine creates a new VM
func (exo *Client) DeployVirtualMachine(req *DeployVirtualMachineRequest, async AsyncInfo) (*VirtualMachine, error) {
	return exo.doVirtualMachine(req, async)
}

// StartVirtualMachine starts the VM and returns its new state
func (exo *Client) StartVirtualMachine(req *StartVirtualMachineRequest, async AsyncInfo) (*VirtualMachine, error) {
	return exo.doVirtualMachine(req, async)
}

// StopVirtualMachine stops the VM and returns its new state
func (exo *Client) StopVirtualMachine(req *StopVirtualMachineRequest, async AsyncInfo) (*VirtualMachine, error) {
	return exo.doVirtualMachine(req, async)
}

// RebootVirtualMachine reboots the VM and returns its new state
func (exo *Client) RebootVirtualMachine(req *RebootVirtualMachineRequest, async AsyncInfo) (*VirtualMachine, error) {
	return exo.doVirtualMachine(req, async)
}

// DestroyVirtualMachine destroy the VM
func (exo *Client) DestroyVirtualMachine(req *DestroyVirtualMachineRequest, async AsyncInfo) (*VirtualMachine, error) {
	return exo.doVirtualMachine(req, async)
}

// doVirtualMachine is a utility function to perform the API call
func (exo *Client) doVirtualMachine(command Request, async AsyncInfo) (*VirtualMachine, error) {
	var r VirtualMachineResponse
	err := exo.AsyncRequest(command, &r, async)
	if err != nil {
		return nil, err
	}

	return r.VirtualMachine, nil
}

// GetVirtualMachine
//
/*
func (exo *Client) GetVirtualMachine(virtualMachineId string) (*VirtualMachine, error) {

	params := url.Values{}
	params.Set("id", virtualMachineId)

	resp, err := exo.Request()
	if err != nil {
		return nil, err
	}

	var r ListVirtualMachinesResponse

	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	if len(r.VirtualMachine) == 1 {
		return r.VirtualMachine[0], nil
	} else {
		return nil, fmt.Errorf("cannot retrieve virtualmachine with id %s", virtualMachineId)
	}
}*/

// ListVirtualMachines lists all the VM
//
// http://cloudstack.apache.org/api/apidocs-4.10/apis/listVirtualMachines.html
func (exo *Client) ListVirtualMachines(req *ListVirtualMachinesRequest) ([]*VirtualMachine, error) {
	var r ListVirtualMachinesResponse
	err := exo.Request(req, &r)
	if err != nil {
		return nil, err
	}

	return r.VirtualMachine, nil
}

// XXX many calls are missing
//
// http://cloudstack.apache.org/api/apidocs-4.10/
