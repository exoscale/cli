package egoscale

func (*ListAccounts) name() string {
	return "listAccounts"
}

func (*ListAccounts) description() string {
	return "Lists accounts and provides detailed account information for listed accounts"
}

func (*ListAccounts) response() interface{} {
	return new(ListAccountsResponse)
}

func (*EnableAccount) name() string {
	return "enableAccount"
}

func (*EnableAccount) description() string {
	return "Enables an account"
}

func (*EnableAccount) response() interface{} {
	return new(Account)
}

func (*DisableAccount) name() string {
	return "disableAccount"
}

func (*DisableAccount) description() string {
	return "Disables an account"
}

func (*DisableAccount) asyncResponse() interface{} {
	return new(Account)
}
