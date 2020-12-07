package v2

//go:generate oapi-codegen -generate types,client -package v2 -o v2.gen.go ../../public-api.openapi.json

// optionalString returns the dereferenced string value of v if not nil, otherwise an empty string.
func optionalString(v *string) string {
	if v != nil {
		return *v
	}

	return ""
}
