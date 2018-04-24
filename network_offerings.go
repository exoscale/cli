package egoscale

// name returns the CloudStack API command name
func (*ListNetworkOfferings) name() string {
	return "listNetworkOfferings"
}

func (*ListNetworkOfferings) response() interface{} {
	return new(ListNetworkOfferingsResponse)
}
