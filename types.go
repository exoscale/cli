package egoscale

import (
	"net/http"
)

type Client struct {
	client    *http.Client
	endpoint  string
	apiKey    string
	apiSecret string
}

type ErrorResponse struct {
	ErrorCode   int      `json:"errorcode"`
	CSErrorCode int      `json:"cserrorcode"`
	ErrorText   string   `json:"errortext"`
	UuidList    []string `json:"uuidlist,omitempty"`
}

type StandardResponse struct {
	Success     string `json:"success"`
	DisplayText string `json:"displaytext"`
}

type Topology struct {
	Zones          map[string]*Zone
	Images         map[string]map[int]string
	Profiles       map[string]string
	Keypairs       []string
	SecurityGroups map[string]string
	AffinityGroups map[string]string
}

// VirtualMachineProfile represents the machine creation request
type VirtualMachineProfile struct {
	Name            string
	SecurityGroups  []string
	Keypair         string
	Userdata        string
	ServiceOffering string
	Template        string
	Zone            string
	AffinityGroups  []string
}

// IpProfile represents the IP creation request
type IpAddressProfile struct {
	Zone string
}

// AsyncInfo represents the details for any async call
type AsyncInfo struct {
	Retries int
	Delay   int
}

type Zone struct {
	Allocationstate       string            `json:"allocationstate,omitempty"`
	Description           string            `json:"description,omitempty"`
	Displaytext           string            `json:"displaytext,omitempty"`
	Domain                string            `json:"domain,omitempty"`
	Domainid              string            `json:"domainid,omitempty"`
	Domainname            string            `json:"domainname,omitempty"`
	Id                    string            `json:"id,omitempty"`
	Internaldns1          string            `json:"internaldns1,omitempty"`
	Internaldns2          string            `json:"internaldns2,omitempty"`
	Ip6dns1               string            `json:"ip6dns1,omitempty"`
	Ip6dns2               string            `json:"ip6dns2,omitempty"`
	Localstorageenabled   bool              `json:"localstorageenabled,omitempty"`
	Name                  string            `json:"name,omitempty"`
	Networktype           string            `json:"networktype,omitempty"`
	Resourcedetails       map[string]string `json:"resourcedetails,omitempty"`
	Securitygroupsenabled bool              `json:"securitygroupsenabled,omitempty"`
	Vlan                  string            `json:"vlan,omitempty"`
	Zonetoken             string            `json:"zonetoken,omitempty"`
}

type ListServiceOfferingsResponse struct {
	Count            int                `json:"count"`
	ServiceOfferings []*ServiceOffering `json:"serviceoffering"`
}

type ServiceOffering struct {
	CpuNumber              int               `json:"cpunumber,omitempty"`
	CpuSpeed               int               `json:"cpuspeed,omitempty"`
	DisplayText            string            `json:"displaytext,omitempty"`
	Domain                 string            `json:"domain,omitempty"`
	DomainId               string            `json:"domainid,omitempty"`
	HostTags               string            `json:"hosttags,omitempty"`
	Id                     string            `json:"id,omitempty"`
	IsCustomized           bool              `json:"iscustomized,omitempty"`
	IsSystem               bool              `json:"issystem,omitempty"`
	IsVolatile             bool              `json:"isvolatile,omitempty"`
	Memory                 int               `json:"memory,omitempty"`
	Name                   string            `json:"name,omitempty"`
	NetworkRate            int               `json:"networkrate,omitempty"`
	ServiceOfferingDetails map[string]string `json:"serviceofferingdetails,omitempty"`
}

type Template struct {
	Account               string            `json:"account,omitempty"`
	Accountid             string            `json:"accountid,omitempty"`
	Bootable              bool              `json:"bootable,omitempty"`
	Checksum              string            `json:"checksum,omitempty"`
	Created               string            `json:"created,omitempty"`
	CrossZones            bool              `json:"crossZones,omitempty"`
	Details               map[string]string `json:"details,omitempty"`
	Displaytext           string            `json:"displaytext,omitempty"`
	Domain                string            `json:"domain,omitempty"`
	Domainid              string            `json:"domainid,omitempty"`
	Format                string            `json:"format,omitempty"`
	Hostid                string            `json:"hostid,omitempty"`
	Hostname              string            `json:"hostname,omitempty"`
	Hypervisor            string            `json:"hypervisor,omitempty"`
	Id                    string            `json:"id,omitempty"`
	Isdynamicallyscalable bool              `json:"isdynamicallyscalable,omitempty"`
	Isextractable         bool              `json:"isextractable,omitempty"`
	Isfeatured            bool              `json:"isfeatured,omitempty"`
	Ispublic              bool              `json:"ispublic,omitempty"`
	Isready               bool              `json:"isready,omitempty"`
	Name                  string            `json:"name,omitempty"`
	Ostypeid              string            `json:"ostypeid,omitempty"`
	Ostypename            string            `json:"ostypename,omitempty"`
	Passwordenabled       bool              `json:"passwordenabled,omitempty"`
	Project               string            `json:"project,omitempty"`
	Projectid             string            `json:"projectid,omitempty"`
	Removed               string            `json:"removed,omitempty"`
	Size                  int64             `json:"size,omitempty"`
	Sourcetemplateid      string            `json:"sourcetemplateid,omitempty"`
	Sshkeyenabled         bool              `json:"sshkeyenabled,omitempty"`
	Status                string            `json:"status,omitempty"`
	Zoneid                string            `json:"zoneid,omitempty"`
	Zonename              string            `json:"zonename,omitempty"`
}

type ListSSHKeyPairsResponse struct {
	Count       int           `json:"count"`
	SSHKeyPairs []*SSHKeyPair `json:"sshkeypair"`
}

type SSHKeyPair struct {
	Fingerprint string `json:"fingerprint,omitempty"`
	Name        string `json:"name,omitempty"`
}

type ListAffinityGroupsResponse struct {
	Count          int              `json:"count"`
	AffinityGroups []*AffinityGroup `json:"affinitygroup"`
}

type AffinityGroup struct {
	Id                string   `json:"id,omitempty"`
	Account           string   `json:"account,omitempty"`
	Description       string   `json:"description,omitempty"`
	Domain            string   `json:"domain,omitempty"`
	DomainId          string   `json:"domainid,omitempty"`
	Name              string   `json:"name,omitempty"`
	Type              string   `json:"type,omitempty"`
	VirtualMachineIds []string `json:"virtualmachineIds,omitempty"` // *I*ds is not a typo
}

type ListSecurityGroupsResponse struct {
	Count          int              `json:"count"`
	SecurityGroups []*SecurityGroup `json:"securitygroup"`
}

// SecurityGroup represent a firewalling set of rules
type SecurityGroup struct {
	Account      string               `json:"account,omitempty"`
	Description  string               `json:"description,omitempty"`
	Domain       string               `json:"domain,omitempty"`
	Domainid     string               `json:"domainid,omitempty"`
	Id           string               `json:"id,omitempty"`
	Name         string               `json:"name,omitempty"`
	Project      string               `json:"project,omitempty"`
	Projectid    string               `json:"projectid,omitempty"`
	IngressRules []*SecurityGroupRule `json:"ingressrule,omitempty"`
	EgressRules  []*SecurityGroupRule `json:"egressrule,omitempty"`
	Tags         []string             `json:"tags,omitempty"`
}

// SecurityGroupRule represents the ingress/egress rule
type SecurityGroupRule struct {
	Account               string   `json:"account,omitempty"`
	RuleId                string   `json:"ruleid,omitempty"`
	Cidr                  string   `json:"cidr,omitempty"`
	IcmpType              int      `json:"icmptype,omitempty"`
	IcmpCode              int      `json:"icmpcode,omitempty"`
	StartPort             int      `json:"startport,omitempty"`
	EndPort               int      `json:"endport,omitempty"`
	Protocol              string   `json:"protocol,omitempty"`
	Tags                  []string `json:"tags,omitempty"`
	SecurityGroupId       string
	UserSecurityGroupList []*UserSecurityGroup `json:"usersecuritygrouplist,omitempty"`
	JobId                 string               `json:"jobid,omitempty"`
	JobStatus             JobStatusType        `json:"jobstatus,omitempty"`
}

// UserSecurityGroup represents the traffic of another security group
type UserSecurityGroup struct {
	Group   string `json:"group,omitempty"`
	Account string `json:"account,omitempty"`
}

type CreateSecurityGroupResponseWrapper struct {
	Wrapped CreateSecurityGroupResponse `json:"securitygroup"`
}

type CreateSecurityGroupResponse struct {
	Account     string `json:"account,omitempty"`
	Description string `json:"description,omitempty"`
	Domain      string `json:"domain,omitempty"`
	Domainid    string `json:"domainid,omitempty"`
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Project     string `json:"project,omitempty"`
	Projectid   string `json:"projectid,omitempty"`
}

type AuthorizeSecurityGroupIngressResponse struct {
	JobID             string `json:"jobid,omitempty"`
	Account           string `json:"account,omitempty"`
	Cidr              string `json:"cidr,omitempty"`
	Endport           int    `json:"endport,omitempty"`
	Icmpcode          int    `json:"icmpcode,omitempty"`
	Icmptype          int    `json:"icmptype,omitempty"`
	Protocol          string `json:"protocol,omitempty"`
	Ruleid            string `json:"ruleid,omitempty"`
	Securitygroupname string `json:"securitygroupname,omitempty"`
	Startport         int    `json:"startport,omitempty"`
}

type ListVirtualMachinesResponse struct {
	Count           int               `json:"count"`
	VirtualMachines []*VirtualMachine `json:"virtualmachine"`
}

// VirtualMachine reprents a virtual machine
type VirtualMachine struct {
	Id                    string            `json:"id,omitempty"`
	Account               string            `json:"account,omitempty"`
	CpuNumber             int               `json:"cpunumber,omitempty"`
	CpuSpeed              int               `json:"cpuspeed,omitempty"`
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
	InstanceName          string            `json:"instancename,omitempty"`
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
	Password              string            `json:"password,omitempty"`
	PasswordEnabled       bool              `json:"passwordenabled,omitempty"`
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
	UserId                string            `json:"userid,omitempty"`
	UserName              string            `json:"username,omitempty"`
	Vgpu                  string            `json:"vgpu,omitempty"`
	ZoneId                string            `json:"zoneid,omitempty"`
	ZoneName              string            `json:"zonename,omitempty"`
	AffinityGroup         []AffinityGroup   `json:"affinitygroup,omitempty"`
	Nic                   []Nic             `json:"nic,omitempty"`
	SecurityGroup         []SecurityGroup   `json:"securitygroup,omitempty"`
	JobID                 string            `json:"jobid,omitempty"`
	JobStatus             JobStatusType     `json:"jobstatus,omitempty"`
}

// Nic represents a Network Interface Controller (NIC)
type Nic struct {
	Id           string `json:"id,omitempty"`
	BroadcastUri string `json:"broadcasturi,omitempty"`
	Gateway      string `json:"gateway,omitempty"`
	Ip6Address   string `json:"ip6address,omitempty"`
	Ip6Cidr      string `json:"ip6cidr,omitempty"`
	Ip6Gateway   string `json:"ip6gateway,omitempty"`
	IpAddress    string `json:"ipaddress,omitempty"`
	IsDefault    bool   `json:"isdefault,omitempty"`
	IsolationUri string `json:"isolationuri,omitempty"`
	MacAddress   string `json:"macaddress,omitempty"`
	Netmask      string `json:"netmask,omitempty"`
	NetworkId    string `json:"networkid,omitempty"`
	NetworkName  string `json:"networkname,omitempty"`
	Secondaryip  []struct {
		Id        string `json:"id,omitempty"`
		IpAddress string `json:"ipaddress,omitempty"`
	} `json:"secondaryip,omitempty"`
	Traffictype      string `json:"traffictype,omitempty"`
	Type             string `json:"type,omitempty"`
	VirtualMachineId string `json:"virtualmachineid,omitempty"`
}

type StartVirtualMachineResponse struct {
	JobID string `json:"jobid,omitempty"`
}

type StopVirtualMachineResponse struct {
	JobID string `json:"jobid,omitempty"`
}

type DestroyVirtualMachineResponse struct {
	JobID string `json:"jobid,omitempty"`
}

type CreateSSHKeyPairWrappedResponse struct {
	Wrapped CreateSSHKeyPairResponse `json:"keypair,omitempty"`
}

type CreateSSHKeyPairResponse struct {
	Privatekey string `json:"privatekey,omitempty"`
}

type DeleteSSHKeyPairResponse struct {
	Privatekey string `json:"privatekey,omitempty"`
}

type DNSDomain struct {
	Id             int64  `json:"id"`
	UserId         int64  `json:"user_id"`
	RegistrantId   int64  `json:"registrant_id,omitempty"`
	Name           string `json:"name"`
	UnicodeName    string `json:"unicode_name"`
	Token          string `json:"token"`
	State          string `json:"state"`
	Language       string `json:"language,omitempty"`
	Lockable       bool   `json:"lockable"`
	AutoRenew      bool   `json:"auto_renew"`
	WhoisProtected bool   `json:"whois_protected"`
	RecordCount    int64  `json:"record_count"`
	ServiceCount   int64  `json:"service_count"`
	ExpiresOn      string `json:"expires_on,omitempty"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type DNSDomainCreateRequest struct {
	Domain struct {
		Name string `json:"name"`
	} `json:"domain"`
}

type DNSRecord struct {
	Id         int64  `json:"id,omitempty"`
	DomainId   int64  `json:"domain_id,omitempty"`
	Name       string `json:"name"`
	Ttl        int    `json:"ttl,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
	Content    string `json:"content"`
	RecordType string `json:"record_type"`
	Prio       int    `json:"prio,omitempty"`
}

type DNSRecordResponse struct {
	Record DNSRecord `json:"record"`
}

type DNSError struct {
	Name []string `json:"name"`
}

// IpAddress represents an IP Address
type IpAddress struct {
	Id                        string        `json:"id"`
	Account                   string        `json:"account,omitempty"`
	AllocatedAt               string        `json:"allocated,omitempty"`
	AssociatedNetworkId       string        `json:"associatednetworkid,omitempty"`
	AssociatedNetworkName     string        `json:"associatednetworkname,omitempty"`
	DomainId                  string        `json:"domainid,omitempty"`
	DomainName                string        `json:"domainname,omitempty"`
	ForDisplay                bool          `json:"fordisplay,omitempty"`
	ForVirtualNetwork         bool          `json:"forvirtualnetwork,omitempty"`
	IpAddress                 string        `json:"ipaddress"`
	IsElastic                 bool          `json:"iselastic,omitempty"`
	IsPortable                bool          `json:"isportable,omitempty"`
	IsSourceNat               bool          `json:"issourcenat,omitempty"`
	IsSystem                  bool          `json:"issystem,omitempty"`
	NetworkId                 string        `json:"networkid,omitempty"`
	PhysicalNetworkId         string        `json:"physicalnetworkid,omitempty"`
	Project                   string        `json:"project,omitempty"`
	ProjectId                 string        `json:"projectid,omitempty"`
	Purpose                   string        `json:"purpose,omitempty"`
	State                     string        `json:"state,omitempty"`
	VirtualMachineDisplayName string        `json:"virtualmachinedisplayname,omitempty"`
	VirtualMachineId          string        `json:"virtualmachineid,omitempty"`
	VirtualMachineName        string        `json:"virtualmachineName,omitempty"`
	VlanId                    string        `json:"vlanid,omitempty"`
	VlanName                  string        `json:"vlanname,omitempty"`
	VmIpAddress               string        `json:"vmipaddress,omitempty"`
	VpcId                     string        `json:"vpcid,omitempty"`
	ZoneId                    string        `json:"zoneid,omitempty"`
	ZoneName                  string        `json:"zonename,omitempty"`
	Tags                      []string      `json:"tags,omitempty"`
	JobId                     string        `json:"jobid,omitempty"`
	JobStatus                 JobStatusType `json:"jobstatus,omitempty"`
}

// NicSecondaryIp represents a link between NicId and IpAddress.
type NicSecondaryIp struct {
	Id               string `json:"id"`
	IpAddress        string `json:"ipaddress"`
	NetworkId        string `json:"networkid"`
	NicId            string `json:"nicid"`
	VirtualMachineId string `json:"virtualmachineid,omitempty"`
}
