package egoscale

// APIName returns the CloudStack API command name
func (*ListResourceDetails) APIName() string {
	return "listResourceDetails"
}

func (*ListResourceDetails) response() interface{} {
	return new(ListResourceDetailsResponse)
}
