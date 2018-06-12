package egoscale

func (*ListAPIs) name() string {
	return "listApis"
}

func (*ListAPIs) description() string {
	return "lists all available apis on the server, provided by the Api Discovery plugin"
}

func (*ListAPIs) response() interface{} {
	return new(ListAPIsResponse)
}
