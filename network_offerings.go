package egoscale

func (*ListNetworkOfferings) name() string {
	return "listNetworkOfferings"
}

func (*ListNetworkOfferings) description() string {
	return "Lists all available network offerings."
}

func (*ListNetworkOfferings) response() interface{} {
	return new(ListNetworkOfferingsResponse)
}

func (*UpdateNetworkOffering) name() string {
	return "updateNetworkOffering"
}

func (*UpdateNetworkOffering) description() string {
	return "Updates a network offering."
}

func (*UpdateNetworkOffering) response() interface{} {
	return new(NetworkOffering)
}
