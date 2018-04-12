package egoscale

// ListRequest builds the ListZones request
func (zone *Zone) ListRequest() (ListCommand, error) {
	req := &ListZones{
		DomainID: zone.DomainID,
		ID:       zone.ID,
		Name:     zone.Name,
	}

	return req, nil
}

// APIName returns the CloudStack API command name
func (*ListZones) APIName() string {
	return "listZones"
}

func (*ListZones) response() interface{} {
	return new(ListZonesResponse)
}

// SetPage sets the current page
func (ls *ListZones) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListZones) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

func (*ListZones) each(resp interface{}, callback IterateItemFunc) {
	zones := resp.(*ListZonesResponse)
	for i := range zones.Zone {
		if !callback(&zones.Zone[i], nil) {
			break
		}
	}
}
