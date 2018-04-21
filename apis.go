package egoscale

// APIName returns the CloudStack API command name
func (*ListAPIs) APIName() string {
	return "listApis"
}

func (*ListAPIs) response() interface{} {
	return new(ListAPIsResponse)
}
