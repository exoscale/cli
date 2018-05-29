package egoscale

func (*ListHosts) name() string {
	return "listHosts"
}

func (*ListHosts) response() interface{} {
	return new(ListHostsResponse)
}
