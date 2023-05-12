package cmd

import (
	"fmt"
	"net"
	"strconv"

	"github.com/spf13/cobra"
)

// nilValue is returned by a flag when the value is not set
const nilValue = "nil"

// XXX use reflect to factor those out.
//     e.g. getCustomFlag(cmd *cobra.Command, name string, out interface{}) error {}

type int64PtrValue struct {
	*int64
}

func (v *int64PtrValue) Set(val string) error {
	r, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return err
	}
	v.int64 = &r
	return nil
}

func (v *int64PtrValue) Type() string {
	return "int64"
}

func (v *int64PtrValue) String() string {
	if v.int64 == nil {
		return nilValue
	}
	return strconv.FormatInt(*v.int64, 10)
}

func getInt64CustomFlag(cmd *cobra.Command, name string) (int64PtrValue, error) {
	it := cmd.Flags().Lookup(name)
	if it != nil {
		r := it.Value.(*int64PtrValue)
		if r != nil {
			return *r, nil
		}
	}
	return int64PtrValue{}, fmt.Errorf("unable to get flag %q", name)
}

// ip flag
type ipValue struct {
	IP *net.IP
}

func (v *ipValue) Set(val string) error {
	if val == "" {
		return nil
	}

	ip := net.ParseIP(val)
	if ip == nil {
		return fmt.Errorf("no a valid IP address: got %q", val)
	}

	v.IP = &ip

	return nil
}

func (v *ipValue) Value() net.IP {
	if v.IP == nil {
		return net.IP{}
	}

	return *v.IP
}

func (v *ipValue) Type() string {
	return "IP"
}

func (v *ipValue) String() string {
	if v.IP == nil || *v.IP == nil {
		return nilValue
	}

	return v.IP.String()
}

// getIPValue finds the value of a command by name
func getIPValue(cmd *cobra.Command, name string) (*ipValue, error) {
	it := cmd.Flags().Lookup(name)
	if it != nil {
		r := it.Value.(*ipValue)
		if r != nil {
			return r, nil
		}
	}
	return nil, fmt.Errorf("unable to get flag %q", name)
}
