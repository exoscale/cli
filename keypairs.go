/*
SSH Key Pairs

In addition to username and password (disabled on Exoscale), SSH keys are used to log into the infrastructure.

See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/stable/virtual_machines.html#creating-the-ssh-keypair
*/
package egoscale

// SshKeyPair represents an SSH key pair
type SshKeyPair struct {
	Account     string `json:"account,omitempty"`
	DomainId    string `json:"domainid,omitempty"`
	ProjectId   string `json:"projectid,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Name        string `json:"name,omitempty"`
	PrivateKey  string `json:"privatekey,omitempty"`
}

// CreateSshKeyPairRequest represents a new keypair to be created
//
// http://cloudstack.apache.org/api/apidocs-4.10/apis/createSSHKeyPair.html
type CreateSshKeyPairRequest struct {
	Name      string `json:"name"`
	Account   string `json:"account,omitempty"`
	DomainId  string `json:"domainid,omitempty"`
	ProjectId string `json:"projectid,omitempty"`
}

// Command returns the CloudStack API command
func (req *CreateSshKeyPairRequest) Command() string {
	return "createSSHKeyPair"
}

// CreateSshKeyPairResponse represents the creation of an SSH Key Pair
type CreateSshKeyPairResponse struct {
	KeyPair *SshKeyPair `json:"keypair"`
}

// DeleteSshKeyPairRequest represents a new keypair to be created
//
// http://cloudstack.apache.org/api/apidocs-4.10/apis/deleteSSHKeyPair.html
type DeleteSshKeyPairRequest struct {
	Name      string `json:"name"`
	Account   string `json:"account,omitempty"`
	DomainId  string `json:"domainid,omitempty"`
	ProjectId string `json:"projectid,omitempty"`
}

// Command returns the CloudStack API command
func (req *DeleteSshKeyPairRequest) Command() string {
	return "deleteSSHKeyPair"
}

// SshKeyPairRequest represents a new registration of a public key in a keypair
//
// http://cloudstack.apache.org/api/apidocs-4.10/apis/registerSSHKeyPair.html
type RegisterSshKeyPairRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"publickey"`
	Account   string `json:"account,omitempty"`
	DomainId  string `json:"domainid,omitempty"`
	ProjectId string `json:"projectid,omitempty"`
}

// Command returns the CloudStack API command
func (req *RegisterSshKeyPairRequest) Command() string {
	return "registerSSHKeyPair"
}

// RegisterSshKeyPairResponse represents the creation of an SSH Key Pair
type RegisterSshKeyPairResponse struct {
	KeyPair *SshKeyPair `json:"keypair"`
}

// ListSshKeyPairsRequest represents a query for a list of SSH KeyPairs
//
// http://cloudstack.apache.org/api/apidocs-4.10/apis/listSSHKeyPairs.html
type ListSshKeyPairsRequest struct {
	Account     string `json:"account,omitempty"`
	DomainId    string `json:"domainid,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	IsRecursive bool   `json:"isrecursive,omitempty"`
	Keyword     string `json:"keyword,omitempty"`
	ListAll     bool   `json:"listall,omitempty"`
	Name        string `json:"name,omitempty"`
	Page        string `json:"page,omitempty"`
	PageSize    string `json:"pagesize,omitempty"`
	ProjectId   string `json:"projectid,omitempty"`
}

// Command returns the CloudStack API command
func (req *ListSshKeyPairsRequest) Command() string {
	return "listSSHKeyPairs"
}

// ListSshKeyPairsResponse
type ListSshKeyPairsResponse struct {
	Count      int           `json:"count"`
	SshKeyPair []*SshKeyPair `json:"sshkeypair"`
}

// XXX ResetSshKeyForVirtualMachine
//
// http://cloudstack.apache.org/api/apidocs-4.10/apis/resetSSHKeyForVirtualMachine.html

// Deprecated: CreateKeypair create a new SSH Key Pair
func (exo *Client) CreateKeypair(name string) (*SshKeyPair, error) {
	req := &CreateSshKeyPairRequest{
		Name: name,
	}
	r := new(CreateSshKeyPairResponse)
	err := exo.Request(req, r)
	if err != nil {
		return nil, err
	}

	return r.KeyPair, nil
}

// Deprecated: DeleteKeypair deletes an SSH key pair
func (exo *Client) DeleteKeypair(name string) error {
	req := &DeleteSshKeyPairRequest{
		Name: name,
	}
	return exo.BooleanRequest(req)
}

// RegisterKeypair registers a public key in a keypair
func (exo *Client) RegisterKeypair(name string, publicKey string) (*SshKeyPair, error) {
	req := &RegisterSshKeyPairRequest{
		Name:      name,
		PublicKey: publicKey,
	}
	r := new(RegisterSshKeyPairResponse)
	err := exo.Request(req, r)
	if err != nil {
		return nil, err
	}

	return r.KeyPair, nil
}
