package egoscale

import (
	"context"
	"fmt"

	"github.com/jinzhu/copier"
)

// Get fetches the resource
func (ipaddress *IPAddress) Get(ctx context.Context, client *Client) error {
	if ipaddress.ID == "" && ipaddress.IPAddress == nil {
		return fmt.Errorf("An IPAddress may only be searched using ID or IPAddress")
	}

	req := &ListPublicIPAddresses{
		ID:        ipaddress.ID,
		IPAddress: ipaddress.IPAddress,
		Account:   ipaddress.Account,
		DomainID:  ipaddress.DomainID,
		ProjectID: ipaddress.ProjectID,
		ZoneID:    ipaddress.ZoneID,
	}

	if ipaddress.IsElastic {
		req.IsElastic = &(ipaddress.IsElastic)
	}

	resp, err := client.RequestWithContext(ctx, req)
	if err != nil {
		return err
	}

	ips := resp.(*ListPublicIPAddressesResponse)
	count := len(ips.PublicIPAddress)
	if count == 0 {
		return &ErrorResponse{
			ErrorCode: ParamError,
			ErrorText: fmt.Sprintf("PublicIPAddress not found. id: %s, ipaddress: %s", ipaddress.ID, ipaddress.IPAddress),
		}
	} else if count > 1 {
		return fmt.Errorf("More than one PublicIPAddress was found")
	}

	return copier.Copy(ipaddress, ips.PublicIPAddress[0])
}

// Delete removes the resource
func (ipaddress *IPAddress) Delete(ctx context.Context, client *Client) error {
	if ipaddress.ID == "" {
		return fmt.Errorf("An IPAddress may only be deleted using ID")
	}

	return client.BooleanRequestWithContext(ctx, &DisassociateIPAddress{
		ID: ipaddress.ID,
	})
}

// ResourceType returns the type of the resource
func (*IPAddress) ResourceType() string {
	return "PublicIpAddress"
}

// APIName returns the CloudStack API command name
func (*AssociateIPAddress) APIName() string {
	return "associateIpAddress"
}

func (*AssociateIPAddress) asyncResponse() interface{} {
	return new(AssociateIPAddressResponse)
}

// APIName returns the CloudStack API command name
func (*DisassociateIPAddress) APIName() string {
	return "disassociateIpAddress"
}
func (*DisassociateIPAddress) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// APIName returns the CloudStack API command name
func (*UpdateIPAddress) APIName() string {
	return "updateIpAddress"
}
func (*UpdateIPAddress) asyncResponse() interface{} {
	return new(UpdateIPAddressResponse)
}

// APIName returns the CloudStack API command name
func (*ListPublicIPAddresses) APIName() string {
	return "listPublicIpAddresses"
}

func (*ListPublicIPAddresses) response() interface{} {
	return new(ListPublicIPAddressesResponse)
}
