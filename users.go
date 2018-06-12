package egoscale

func (*RegisterUserKeys) name() string {
	return "registerUserKeys"
}

func (*RegisterUserKeys) description() string {
	return "This command allows a user to register for the developer API, returning a secret key and an API key. This request is made through the integration API port, so it is a privileged command and must be made on behalf of a user. It is up to the implementer just how the username and password are entered, and then how that translates to an integration API request. Both secret key and API key should be returned to the user"
}

func (*RegisterUserKeys) response() interface{} {
	return new(User)
}

func (*CreateUser) name() string {
	return "createUser"
}

func (*CreateUser) description() string {
	return "Creates a user for an account that already exists"
}

func (*CreateUser) response() interface{} {
	return new(User)
}

func (*UpdateUser) name() string {
	return "updateUser"
}

func (*UpdateUser) description() string {
	return "Updates a user account"
}

func (*UpdateUser) response() interface{} {
	return new(User)
}

func (*ListUsers) name() string {
	return "listUsers"
}

func (*ListUsers) description() string {
	return "Lists user accounts"
}

func (*ListUsers) response() interface{} {
	return new(ListUsersResponse)
}

func (*DeleteUser) name() string {
	return "deleteUser"
}

func (*DeleteUser) description() string {
	return "Deletes a user for an account"
}

func (*DeleteUser) response() interface{} {
	return new(booleanResponse)
}
