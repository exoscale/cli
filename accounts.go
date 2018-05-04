package egoscale

func (*ListAccounts) name() string {
	return "listAccounts"
}

func (*ListAccounts) response() interface{} {
	return new(ListAccountsResponse)
}

func (*EnableAccount) name() string {
	return "enableAccount"
}

func (*EnableAccount) response() interface{} {
	return new(Account)
}

func (*DisableAccount) name() string {
	return "disableAccount"
}

func (*DisableAccount) asyncResponse() interface{} {
	return new(Account)
}
