/*

Package admin contains the privileged API calls.

Some aspects of the Exoscale API are restricted to admin privileges, meaning not all field would be useful for everyone if we allowed them to coexist with the base structs. E.g. an admin can list all resources belonging to anyone using the listall=true parameter.

*/
package admin
