package egoscale

func (*RegisterUserKeys) name() string {
	return "registerUserKeys"
}

func (*RegisterUserKeys) response() interface{} {
	return new(User)
}

func (*CreateUser) name() string {
	return "createUser"
}

func (*CreateUser) response() interface{} {
	return new(User)
}

func (*UpdateUser) name() string {
	return "updateUser"
}

func (*UpdateUser) response() interface{} {
	return new(User)
}

func (*ListUsers) name() string {
	return "listUsers"
}

func (*ListUsers) response() interface{} {
	return new(ListUsersResponse)
}

func (*DeleteUser) name() string {
	return "deleteUser"
}

func (*DeleteUser) response() interface{} {
	return new(booleanResponse)
}
