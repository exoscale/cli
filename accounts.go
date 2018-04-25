package egoscale

// name returns the CloudStack API command name
func (*ListAccounts) name() string {
	return "listAccounts"
}

func (*ListAccounts) response() interface{} {
	return new(ListAccountsResponse)
}
