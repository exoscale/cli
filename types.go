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

type Topology struct {
	Zones          map[string]string
	Images         map[string]map[int64]string
	Profiles       map[string]string
	Keypairs       []string
	SecurityGroups map[string]string
	AffinityGroups map[string]string
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
