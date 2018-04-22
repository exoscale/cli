package egoscale

// APIName returns the CloudStack API command name
func (*RegisterUserKeys) APIName() string {
	return "registerUserKeys"
}

func (*RegisterUserKeys) response() interface{} {
	return new(RegisterUserKeysResponse)
}

// APIName returns the CloudStack API command name
func (*CreateUser) APIName() string {
	return "createUser"
}

func (*CreateUser) response() interface{} {
	return new(CreateUserResponse)
}

// APIName returns the CloudStack API command name
func (*UpdateUser) APIName() string {
	return "updateUser"
}

func (*UpdateUser) response() interface{} {
	return new(UpdateUserResponse)
}

// APIName returns the CloudStack API command name
func (*ListUsers) APIName() string {
	return "listUsers"
}

func (*ListUsers) response() interface{} {
	return new(ListUsersResponse)
}
