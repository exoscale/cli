package egoscale

import (
	"net"
)

// VirtualMachine represents a virtual machine
//
// See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/stable/virtual_machines.html
type VirtualMachine struct {
	ID                    string            `json:"id,omitempty"`
	Account               string            `json:"account,omitempty"`
	ClusterID             string            `json:"clusterid,omitempty"`
	ClusterName           string            `json:"clustername,omitempty"`
	CPUNumber             int64             `json:"cpunumber,omitempty"`
	CPUSpeed              int64             `json:"cpuspeed,omitempty"`
	CPUUsed               string            `json:"cpuused,omitempty"`
	Created               string            `json:"created,omitempty"`
	Details               map[string]string `json:"details,omitempty"`
	DiskIoRead            int64             `json:"diskioread,omitempty"`
	DiskIoWrite           int64             `json:"diskiowrite,omitempty"`
	DiskKbsRead           int64             `json:"diskkbsread,omitempty"`
	DiskKbsWrite          int64             `json:"diskkbswrite,omitempty"`
	DiskOfferingID        string            `json:"diskofferingid,omitempty"`
	DiskOfferingName      string            `json:"diskofferingname,omitempty"`
	DisplayName           string            `json:"displayname,omitempty"`
	DisplayVM             bool              `json:"displayvm,omitempty"`
	Domain                string            `json:"domain,omitempty"`
	DomainID              string            `json:"domainid,omitempty"`
	ForVirtualNetwork     bool              `json:"forvirtualnetwork,omitempty"`
	Group                 string            `json:"group,omitempty"`
	GroupID               string            `json:"groupid,omitempty"`
	GuestOsID             string            `json:"guestosid,omitempty"`
	HAEnable              bool              `json:"haenable,omitempty"`
	HostID                string            `json:"hostid,omitempty"`
	HostName              string            `json:"hostname,omitempty"`
	Hypervisor            string            `json:"hypervisor,omitempty"`
	InstanceName          string            `json:"instancename,omitempty"` // root only
	IsDynamicallyScalable bool              `json:"isdynamicallyscalable,omitempty"`
	IsoDisplayText        string            `json:"isodisplaytext,omitempty"`
	IsoID                 string            `json:"isoid,omitempty"`
	IsoName               string            `json:"isoname,omitempty"`
	KeyPair               string            `json:"keypair,omitempty"`
	Memory                int64             `json:"memory,omitempty"`
	MemoryIntFreeKbs      int64             `json:"memoryintfreekbs,omitempty"`
	MemoryKbs             int64             `json:"memorykbs,omitempty"`
	MemoryTargetKbs       int64             `json:"memorytargetkbs,omitempty"`
	Name                  string            `json:"name,omitempty"`
	NetworkKbsRead        int64             `json:"networkkbsread,omitempty"`
	NetworkKbsWrite       int64             `json:"networkkbswrite,omitempty"`
	OsCategoryID          string            `json:"oscategoryid,omitempty"`
	OsTypeID              int64             `json:"ostypeid,omitempty"`
	Password              string            `json:"password,omitempty"`
	PasswordEnabled       bool              `json:"passwordenabled,omitempty"`
	PCIDevices            string            `json:"pcidevices,omitempty"` // not in the doc
	PodID                 string            `json:"podid,omitempty"`
	PodName               string            `json:"podname,omitempty"`
	Project               string            `json:"project,omitempty"`
	ProjectID             string            `json:"projectid,omitempty"`
	PublicIP              string            `json:"publicip,omitempty"`
	PublicIPID            string            `json:"publicipid,omitempty"`
	RootDeviceID          int64             `json:"rootdeviceid,omitempty"`
	RootDeviceType        string            `json:"rootdevicetype,omitempty"`
	ServiceOfferingID     string            `json:"serviceofferingid,omitempty"`
	ServiceOfferingName   string            `json:"serviceofferingname,omitempty"`
	ServiceState          string            `json:"servicestate,omitempty"`
	State                 string            `json:"state,omitempty"`
	TemplateDisplayText   string            `json:"templatedisplaytext,omitempty"`
	TemplateID            string            `json:"templateid,omitempty"`
	TemplateName          string            `json:"templatename,omitempty"`
	UserID                string            `json:"userid,omitempty"`   // not in the doc
	UserName              string            `json:"username,omitempty"` // not in the doc
	Vgpu                  string            `json:"vgpu,omitempty"`     // not in the doc
	ZoneID                string            `json:"zoneid,omitempty"`
	ZoneName              string            `json:"zonename,omitempty"`
	AffinityGroup         []AffinityGroup   `json:"affinitygroup,omitempty"`
	Nic                   []Nic             `json:"nic,omitempty"`
	SecurityGroup         []SecurityGroup   `json:"securitygroup,omitempty"`
	Tags                  []ResourceTag     `json:"tags,omitempty"`
	JobID                 string            `json:"jobid,omitempty"`
	JobStatus             JobStatusType     `json:"jobstatus,omitempty"`
}

// IPToNetwork represents a mapping between ip and networks
type IPToNetwork struct {
	IP        string `json:"ip,omitempty"`
	IPV6      string `json:"ipv6,omitempty"`
	NetworkID string `json:"networkid,omitempty"`
}

// Password represents an encrypted password
//
// TODO: method to decrypt it, https://cwiki.apache.org/confluence/pages/viewpage.action?pageId=34014652
type Password struct {
	EncryptedPassword string `json:"encryptedpassword"`
}

// VirtualMachineResponse represents a generic Virtual Machine response
type VirtualMachineResponse struct {
	VirtualMachine VirtualMachine `json:"virtualmachine"`
}

// DeployVirtualMachine (Async) represents the machine creation
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/deployVirtualMachine.html
type DeployVirtualMachine struct {
	ServiceOfferingID  string            `json:"serviceofferingid"`
	TemplateID         string            `json:"templateid"`
	ZoneID             string            `json:"zoneid"`
	Account            string            `json:"account,omitempty"`
	AffinityGroupIDs   []string          `json:"affinitygroupids,omitempty"`   // mutually exclusive with AffinityGroupNames
	AffinityGroupNames []string          `json:"affinitygroupnames,omitempty"` // mutually exclusive with AffinityGroupIDs
	CustomID           string            `json:"customid,omitempty"`           // root only
	DeploymentPlanner  string            `json:"deploymentplanner,omitempty"`  // root only
	Details            map[string]string `json:"details,omitempty"`
	DiskOfferingID     string            `json:"diskofferingid,omitempty"`
	DisplayName        string            `json:"displayname,omitempty"`
	DisplayVM          *bool             `json:"displayvm,omitempty"`
	DomainID           string            `json:"domainid,omitempty"`
	Group              string            `json:"group,omitempty"`
	HostID             string            `json:"hostid,omitempty"`
	Hypervisor         string            `json:"hypervisor,omitempty"`
	IP4                *bool             `json:"ip4,omitempty"` // Exoscale specific
	IP6                *bool             `json:"ip6,omitempty"` // Exoscale specific
	IPAddress          net.IP            `json:"ipaddress,omitempty"`
	IP6Address         net.IP            `json:"ip6address,omitempty"`
	IPToNetworkList    []IPToNetwork     `json:"iptonetworklist,omitempty"`
	Keyboard           string            `json:"keyboard,omitempty"`
	KeyPair            string            `json:"keypair,omitempty"`
	Name               string            `json:"name,omitempty"`
	NetworkIDs         []string          `json:"networkids,omitempty"` // mutually exclusive with IPToNetworkList
	ProjectID          string            `json:"projectid,omitempty"`
	RootDiskSize       int64             `json:"rootdisksize,omitempty"`       // in GiB
	SecurityGroupIDs   []string          `json:"securitygroupids,omitempty"`   // mutually exclusive with SecurityGroupNames
	SecurityGroupNames []string          `json:"securitygroupnames,omitempty"` // mutually exclusive with SecurityGroupIDs
	Size               string            `json:"size,omitempty"`               // mutually exclusive with DiskOfferingID
	StartVM            *bool             `json:"startvm,omitempty"`
	UserData           string            `json:"userdata,omitempty"` // the client is responsible to base64/gzip it
}

// DeployVirtualMachineResponse represents a deployed VM instance
type DeployVirtualMachineResponse VirtualMachineResponse

// StartVirtualMachine (Async) represents the creation of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/startVirtualMachine.html
type StartVirtualMachine struct {
	ID                string `json:"id"`
	DeploymentPlanner string `json:"deploymentplanner,omitempty"` // root only
	HostID            string `json:"hostid,omitempty"`            // root only
}

// StartVirtualMachineResponse represents a started VM instance
type StartVirtualMachineResponse VirtualMachineResponse

// StopVirtualMachine (Async) represents the stopping of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/stopVirtualMachine.html
type StopVirtualMachine struct {
	ID     string `json:"id"`
	Forced *bool  `json:"forced,omitempty"`
}

// StopVirtualMachineResponse represents a stopped VM instance
type StopVirtualMachineResponse VirtualMachineResponse

// RebootVirtualMachine (Async) represents the rebooting of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/rebootVirtualMachine.html
type RebootVirtualMachine struct {
	ID string `json:"id"`
}

// RestoreVirtualMachine (Async) represents the restoration of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/restoreVirtualMachine.html
type RestoreVirtualMachine struct {
	VirtualMachineID string `json:"virtualmachineid"`
	TemplateID       string `json:"templateid,omitempty"`
	RootDiskSize     string `json:"rootdisksize,omitempty"` // in GiB, Exoscale specific
}

// RestoreVirtualMachineResponse represents a restored VM instance
type RestoreVirtualMachineResponse VirtualMachineResponse

// RecoverVirtualMachine represents the restoration of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/recoverVirtualMachine.html
type RecoverVirtualMachine struct {
	ID string `json:"virtualmachineid"`
}

// RecoverVirtualMachineResponse represents a recovered VM instance
type RecoverVirtualMachineResponse VirtualMachineResponse

// DestroyVirtualMachine (Async) represents the destruction of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/destroyVirtualMachine.html
type DestroyVirtualMachine struct {
	ID      string `json:"id"`
	Expunge *bool  `json:"expunge,omitempty"`
}

// DestroyVirtualMachineResponse represents a destroyed VM instance
type DestroyVirtualMachineResponse VirtualMachineResponse

// UpdateVirtualMachine represents the update of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/updateVirtualMachine.html
type UpdateVirtualMachine struct {
	ID                    string            `json:"id"`
	CustomID              string            `json:"customid,omitempty"` // root only
	Details               map[string]string `json:"details,omitempty"`
	DisplayName           string            `json:"displayname,omitempty"`
	DisplayVM             *bool             `json:"displayvm,omitempty"`
	Group                 string            `json:"group,omitempty"`
	HAEnable              *bool             `json:"haenable,omitempty"`
	IsDynamicallyScalable *bool             `json:"isdynamicallyscalable,omitempty"`
	Name                  string            `json:"name,omitempty"` // must reboot
	OSTypeID              int64             `json:"ostypeid,omitempty"`
	SecurityGroupIDs      []string          `json:"securitygroupids,omitempty"`
	UserData              string            `json:"userdata,omitempty"`
}

// ScaleVirtualMachine (Async) represents the scaling of a VM
//
// ChangeServiceForVirtualMachine does the same thing but returns the
// new Virtual Machine which is more consistent with the rest of the API.
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/scaleVirtualMachine.html
type ScaleVirtualMachine struct {
	ID                string            `json:"id"`
	ServiceOfferingID string            `json:"serviceofferingid"`
	Details           map[string]string `json:"details,omitempty"`
}

// ChangeServiceForVirtualMachine represents the scaling of a VM
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/changeServiceForVirtualMachine.html
type ChangeServiceForVirtualMachine ScaleVirtualMachine

// ChangeServiceForVirtualMachineResponse represents an changed VM instance
type ChangeServiceForVirtualMachineResponse VirtualMachineResponse

// ResetPasswordForVirtualMachine (Async) represents the scaling of a VM
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/resetPasswordForVirtualMachine.html
type ResetPasswordForVirtualMachine ScaleVirtualMachine

// ResetPasswordForVirtualMachineResponse represents the updated vm
type ResetPasswordForVirtualMachineResponse VirtualMachineResponse

// GetVMPassword asks for an encrypted password
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/getVMPassword.html
type GetVMPassword struct {
	ID string `json:"id"`
}

// GetVMPasswordResponse represents the encrypted password
type GetVMPasswordResponse struct {
	// Base64 encrypted password for the VM
	Password Password `json:"password"`
}

// ListVirtualMachines represents a search for a VM
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/listVirtualMachine.html
type ListVirtualMachines struct {
	Account           string        `json:"account,omitempty"`
	AffinityGroupID   string        `json:"affinitygroupid,omitempty"`
	Details           []string      `json:"details,omitempty"`   // default to "all"
	DisplayVM         *bool         `json:"displayvm,omitempty"` // root only
	DomainID          string        `json:"domainid,omitempty"`
	ForVirtualNetwork *bool         `json:"forvirtualnetwork,omitempty"`
	GroupID           string        `json:"groupid,omitempty"`
	HostID            string        `json:"hostid,omitempty"`
	Hypervisor        string        `json:"hypervisor,omitempty"`
	ID                string        `json:"id,omitempty"`
	IDs               []string      `json:"ids,omitempty"` // mutually exclusive with id
	IPAddress         net.IP        `json:"ipaddress,omitempty"`
	IsoID             string        `json:"isoid,omitempty"`
	IsRecursive       *bool         `json:"isrecursive,omitempty"`
	KeyPair           string        `json:"keypair,omitempty"` // not implemented at Exoscale
	Keyword           string        `json:"keyword,omitempty"`
	ListAll           *bool         `json:"listall,omitempty"`
	Name              string        `json:"name,omitempty"`
	NetworkID         string        `json:"networkid,omitempty"`
	Page              int           `json:"page,omitempty"`
	PageSize          int           `json:"pagesize,omitempty"`
	PodID             string        `json:"podid,omitempty"`
	ProjectID         string        `json:"projectid,omitempty"`
	ServiceOfferindID string        `json:"serviceofferingid,omitempty"`
	State             string        `json:"state,omitempty"` // Running, Stopped, Present, ...
	StorageID         string        `json:"storageid,omitempty"`
	Tags              []ResourceTag `json:"tags,omitempty"`
	TemplateID        string        `json:"templateid,omitempty"`
	UserID            string        `json:"userid,omitempty"`
	VpcID             string        `json:"vpcid,omitempty"`
	ZoneID            string        `json:"zoneid,omitempty"`
}

// ListVirtualMachinesResponse represents a list of virtual machines
type ListVirtualMachinesResponse struct {
	Count          int              `json:"count"`
	VirtualMachine []VirtualMachine `json:"virtualmachine"`
}

// AddNicToVirtualMachine (Async) adds a NIC to a VM
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/addNicToVirtualMachine.html
type AddNicToVirtualMachine struct {
	NetworkID        string `json:"networkid"`
	VirtualMachineID string `json:"virtualmachineid"`
	IPAddress        net.IP `json:"ipaddress,omitempty"`
}

// AddNicToVirtualMachineResponse represents the modified VM
type AddNicToVirtualMachineResponse VirtualMachineResponse

// RemoveNicFromVirtualMachine (Async) removes a NIC from a VM
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/removeNicFromVirtualMachine.html
type RemoveNicFromVirtualMachine struct {
	NicID            string `json:"nicid"`
	VirtualMachineID string `json:"virtualmachineid"`
}

// RemoveNicFromVirtualMachineResponse represents the modified VM
type RemoveNicFromVirtualMachineResponse VirtualMachineResponse

// UpdateDefaultNicForVirtualMachine (Async) adds a NIC to a VM
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/updateDefaultNicForVirtualMachine.html
type UpdateDefaultNicForVirtualMachine struct {
	NetworkID        string `json:"networkid"`
	VirtualMachineID string `json:"virtualmachineid"`
	IPAddress        net.IP `json:"ipaddress,omitempty"`
}

// UpdateDefaultNicForVirtualMachineResponse represents the modified VM
type UpdateDefaultNicForVirtualMachineResponse VirtualMachineResponse
