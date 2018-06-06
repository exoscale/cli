package egoscale

import (
	"fmt"
)

// ListRequest builds the ListSecurityGroups request
func (so *ServiceOffering) ListRequest() (ListCommand, error) {
	req := &ListServiceOfferings{
		ID:           so.ID,
		DomainID:     so.DomainID,
		IsSystem:     &so.IsSystem,
		Name:         so.Name,
		Restricted:   &so.Restricted,
		SystemVMType: so.SystemVMType,
	}

	return req, nil
}

// name returns the CloudStack API command name
func (*ListServiceOfferings) name() string {
	return "listServiceOfferings"
}

func (*ListServiceOfferings) response() interface{} {
	return new(ListServiceOfferingsResponse)
}

// SetPage sets the current page
func (lso *ListServiceOfferings) SetPage(page int) {
	lso.Page = page
}

// SetPageSize sets the page size
func (lso *ListServiceOfferings) SetPageSize(pageSize int) {
	lso.PageSize = pageSize
}

func (*ListServiceOfferings) each(resp interface{}, callback IterateItemFunc) {
	sos, ok := resp.(*ListServiceOfferingsResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type. ListServiceOfferingsResponse expected, got %T", resp))
		return
	}

	for i := range sos.ServiceOffering {
		if !callback(&sos.ServiceOffering[i], nil) {
			break
		}
	}
}
