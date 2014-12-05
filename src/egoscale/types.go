package egoscale

import (
	"net/http"
)

type Client struct {
	client *http.Client
	endpoint string
	apiKey string
	apiSecret string
}


type Error struct {
	ErrorCode int `json:"errorcode"`
	CSErrorCode int `json:"cserrorcode"`
	ErrorText string `json:"errortext"`
}

type Topology struct {
	Zones map[string]string
	Images map[string]map[int]string
	Profiles map[string]string
	Keypairs []string
	SecurityGroups map[string]string
}

type ListZonesResponse struct {
	Count int `json:"count"`
	Zones []*Zone `json:"zone"`

}
type Zone struct {
	Allocationstate string `json:"allocationstate,omitempty"`
	Capacity []struct {
		Capacitytotal int `json:"capacitytotal,omitempty"`
		Capacityused int `json:"capacityused,omitempty"`
		Clusterid string `json:"clusterid,omitempty"`
		Clustername string `json:"clustername,omitempty"`
		Percentused string `json:"percentused,omitempty"`
		Podid string `json:"podid,omitempty"`
		Podname string `json:"podname,omitempty"`
		Type int `json:"type,omitempty"`
		Zoneid string `json:"zoneid,omitempty"`
		Zonename string `json:"zonename,omitempty"`

	} `json:"capacity,omitempty"`
	Description string `json:"description,omitempty"`
	Dhcpprovider string `json:"dhcpprovider,omitempty"`
	Displaytext string `json:"displaytext,omitempty"`
	Dns1 string `json:"dns1,omitempty"`
	Dns2 string `json:"dns2,omitempty"`
	Domain string `json:"domain,omitempty"`
	Domainid string `json:"domainid,omitempty"`
	Domainname string `json:"domainname,omitempty"`
	Guestcidraddress string `json:"guestcidraddress,omitempty"`
	Id string `json:"id,omitempty"`
	Internaldns1 string `json:"internaldns1,omitempty"`
	Internaldns2 string `json:"internaldns2,omitempty"`
	Ip6dns1 string `json:"ip6dns1,omitempty"`
	Ip6dns2 string `json:"ip6dns2,omitempty"`
	Localstorageenabled bool `json:"localstorageenabled,omitempty"`
	Name string `json:"name,omitempty"`
	Networktype string `json:"networktype,omitempty"`
	Resourcedetails map[string]string `json:"resourcedetails,omitempty"`
	Securitygroupsenabled bool `json:"securitygroupsenabled,omitempty"`
	Tags []struct {
		Account string `json:"account,omitempty"`
		Customer string `json:"customer,omitempty"`
		Domain string `json:"domain,omitempty"`
		Domainid string `json:"domainid,omitempty"`
		Key string `json:"key,omitempty"`
		Project string `json:"project,omitempty"`
		Projectid string `json:"projectid,omitempty"`
		Resourceid string `json:"resourceid,omitempty"`
		Resourcetype string `json:"resourcetype,omitempty"`
		Value string `json:"value,omitempty"`

	} `json:"tags,omitempty"`
	Vlan string `json:"vlan,omitempty"`
	Zonetoken string `json:"zonetoken,omitempty"`
}

type ListServiceOfferingsResponse struct {
	Count int `json:"count"`
	ServiceOfferings []*ServiceOffering `json:"serviceoffering"`

}
type ServiceOffering struct {
	Cpunumber int `json:"cpunumber,omitempty"`
	Cpuspeed int `json:"cpuspeed,omitempty"`
	Created string `json:"created,omitempty"`
	Defaultuse bool `json:"defaultuse,omitempty"`
	Deploymentplanner string `json:"deploymentplanner,omitempty"`
	DiskBytesReadRate int `json:"diskBytesReadRate,omitempty"`
	DiskBytesWriteRate int `json:"diskBytesWriteRate,omitempty"`
	DiskIopsReadRate int `json:"diskIopsReadRate,omitempty"`
	DiskIopsWriteRate int `json:"diskIopsWriteRate,omitempty"`
	Displaytext string `json:"displaytext,omitempty"`
	Domain string `json:"domain,omitempty"`
	Domainid string `json:"domainid,omitempty"`
	Hosttags string `json:"hosttags,omitempty"`
	Id string `json:"id,omitempty"`
	Iscustomized bool `json:"iscustomized,omitempty"`
	Issystem bool `json:"issystem,omitempty"`
	Isvolatile bool `json:"isvolatile,omitempty"`
	Limitcpuuse bool `json:"limitcpuuse,omitempty"`
	Memory int `json:"memory,omitempty"`
	Name string `json:"name,omitempty"`
	Networkrate int `json:"networkrate,omitempty"`
	Offerha bool `json:"offerha,omitempty"`
	Serviceofferingdetails map[string]string `json:"serviceofferingdetails,omitempty"`
	Storagetype string `json:"storagetype,omitempty"`
	Systemvmtype string `json:"systemvmtype,omitempty"`
	Tags string `json:"tags,omitempty"`

}

type ListTemplatesResponse struct {
	Count int `json:"count"`
	Templates []*Template `json:"template"`
}

type Template struct {
	Account string `json:"account,omitempty"`
	Accountid string `json:"accountid,omitempty"`
	Bootable bool `json:"bootable,omitempty"`
	Checksum string `json:"checksum,omitempty"`
	Created string `json:"created,omitempty"`
	CrossZones bool `json:"crossZones,omitempty"`
	Details map[string]string `json:"details,omitempty"`
	Displaytext string `json:"displaytext,omitempty"`
	Domain string `json:"domain,omitempty"`
	Domainid string `json:"domainid,omitempty"`
	Format string `json:"format,omitempty"`
	Hostid string `json:"hostid,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Hypervisor string `json:"hypervisor,omitempty"`
	Id string `json:"id,omitempty"`
	Isdynamicallyscalable bool `json:"isdynamicallyscalable,omitempty"`
	Isextractable bool `json:"isextractable,omitempty"`
	Isfeatured bool `json:"isfeatured,omitempty"`
	Ispublic bool `json:"ispublic,omitempty"`
	Isready bool `json:"isready,omitempty"`
	Name string `json:"name,omitempty"`
	Ostypeid string `json:"ostypeid,omitempty"`
	Ostypename string `json:"ostypename,omitempty"`
	Passwordenabled bool `json:"passwordenabled,omitempty"`
	Project string `json:"project,omitempty"`
	Projectid string `json:"projectid,omitempty"`
	Removed string `json:"removed,omitempty"`
	Size int `json:"size,omitempty"`
	Sourcetemplateid string `json:"sourcetemplateid,omitempty"`
	Sshkeyenabled bool `json:"sshkeyenabled,omitempty"`
	Status string `json:"status,omitempty"`
	Tags []struct {
		Account string `json:"account,omitempty"`
		Customer string `json:"customer,omitempty"`
		Domain string `json:"domain,omitempty"`
		Domainid string `json:"domainid,omitempty"`
		Key string `json:"key,omitempty"`
		Project string `json:"project,omitempty"`
		Projectid string `json:"projectid,omitempty"`
		Resourceid string `json:"resourceid,omitempty"`
		Resourcetype string `json:"resourcetype,omitempty"`
		Value string `json:"value,omitempty"`
	} `json:"tags,omitempty"`
	Templatetag string `json:"templatetag,omitempty"`
	Templatetype string `json:"templatetype,omitempty"`
	Zoneid string `json:"zoneid,omitempty"`
	Zonename string `json:"zonename,omitempty"`
}

type ListSSHKeyPairsResponse struct {
	Count int `json:"count"`
	SSHKeyPairs []*SSHKeyPair `json:"sshkeypair"`
}

type SSHKeyPair struct {
	Fingerprint string `json:"fingerprint,omitempty"`
	Name string `json:"name,omitempty"`
}


type ListSecurityGroupsResponse struct {
	Count int `json:"count"`
	SecurityGroups []*SecurityGroup `json:"securitygroup"`

}
type SecurityGroup struct {
	Account string `json:"account,omitempty"`
	Description string `json:"description,omitempty"`
	Domain string `json:"domain,omitempty"`
	Domainid string `json:"domainid,omitempty"`
	Egressrule []struct {
		Account string `json:"account,omitempty"`
		Cidr string `json:"cidr,omitempty"`
		Endport int `json:"endport,omitempty"`
		Icmpcode int `json:"icmpcode,omitempty"`
		Icmptype int `json:"icmptype,omitempty"`
		Protocol string `json:"protocol,omitempty"`
		Ruleid string `json:"ruleid,omitempty"`
		Securitygroupname string `json:"securitygroupname,omitempty"`
		Startport int `json:"startport,omitempty"`

	} `json:"egressrule,omitempty"`
	Id string `json:"id,omitempty"`
	Ingressrule []struct {
		Account string `json:"account,omitempty"`
		Cidr string `json:"cidr,omitempty"`
		Endport int `json:"endport,omitempty"`
		Icmpcode int `json:"icmpcode,omitempty"`
		Icmptype int `json:"icmptype,omitempty"`
		Protocol string `json:"protocol,omitempty"`
		Ruleid string `json:"ruleid,omitempty"`
		Securitygroupname string `json:"securitygroupname,omitempty"`
		Startport int `json:"startport,omitempty"`

	} `json:"ingressrule,omitempty"`
	Name string `json:"name,omitempty"`
	Project string `json:"project,omitempty"`
	Projectid string `json:"projectid,omitempty"`
	Tags []struct {
		Account string `json:"account,omitempty"`
		Customer string `json:"customer,omitempty"`
		Domain string `json:"domain,omitempty"`
		Domainid string `json:"domainid,omitempty"`
		Key string `json:"key,omitempty"`
		Project string `json:"project,omitempty"`
		Projectid string `json:"projectid,omitempty"`
		Resourceid string `json:"resourceid,omitempty"`
		Resourcetype string `json:"resourcetype,omitempty"`
		Value string `json:"value,omitempty"`

	} `json:"tags,omitempty"`

}
