package cmd

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type uint8PtrValue struct {
	*uint8
}

func (v *uint8PtrValue) Set(val string) error {
	r, err := strconv.ParseUint(val, 10, 8)
	if err != nil {
		return err
	}
	res := uint8(r)
	v.uint8 = &res
	return nil
}

func (v *uint8PtrValue) Type() string {
	return "uint8"
}

func (v *uint8PtrValue) String() string {
	if v.uint8 == nil {
		return "nil"
	}
	return strconv.FormatUint(uint64(*v.uint8), 10)
}

func getUint8CustomFlag(cmd *cobra.Command, name string) (uint8PtrValue, error) {
	it := cmd.Flags().Lookup(name)
	if it != nil {
		r := it.Value.(*uint8PtrValue)
		if r != nil {
			return *r, nil
		}
	}
	return uint8PtrValue{}, fmt.Errorf("unable to get flag %q", name)
}

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
		return "nil"
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

// EXO API flags

// uuid flag
type uuid struct {
	UUID **egoscale.UUID
}

func (v *uuid) Set(val string) error {
	r, err := egoscale.ParseUUID(val)
	if err != nil {
		return err
	}
	*(v.UUID) = r

	return nil
}

func (v *uuid) Type() string {
	return "UUID"
}

func (v *uuid) String() string {
	if v.UUID == nil || *(v.UUID) == nil {
		return "nil"
	}

	return (*(v.UUID)).String()
}

//cidr flag
type cidr struct {
	CIDR **egoscale.CIDR
}

func (v *cidr) Set(val string) error {
	r, err := egoscale.ParseCIDR(val)
	if err != nil {
		return err
	}
	*(v.CIDR) = r

	return nil
}

func (v *cidr) Type() string {
	return "CIDR"
}

func (v *cidr) String() string {
	if v.CIDR == nil || *(v.CIDR) == nil {
		return "nil"
	}

	return (*(v.CIDR)).String()
}

//bool flag
type boolFlag struct {
	bool **bool
}

func (v *boolFlag) Set(val string) error {

	r := true
	if val == "true" {
		*(v.bool) = &r
		return nil
	}

	r = false

	*(v.bool) = &r

	return nil
}

func (v *boolFlag) Type() string {
	return "bool"
}

func (v *boolFlag) String() string {
	if v.bool == nil || *(v.bool) == nil {
		return "nil"
	}

	if *(*(v.bool)) {
		return "true"
	}
	return "false"
}

// uuid list flag
type uuidListGeneric struct {
	value *[]egoscale.UUID
}

func (g *uuidListGeneric) Set(value string) error {
	m := g.value
	if *m == nil {
		n := make([]egoscale.UUID, 0)
		*m = n
	}

	values := strings.Split(value, ",")

	for _, value := range values {
		uuid, err := egoscale.ParseUUID(value)
		if err != nil {
			return err
		}
		*m = append(*m, *uuid)
	}
	return nil
}

func (g *uuidListGeneric) Type() string {
	return "uuidListGeneric"
}

func (g *uuidListGeneric) String() string {
	m := g.value
	if m == nil || *m == nil {
		return ""
	}
	vs := make([]string, 0, len(*m))
	for _, v := range *m {
		vs = append(vs, v.String())
	}

	return strings.Join(vs, ",")
}

//cidr list generic
type cidrListGeneric struct {
	value *[]egoscale.CIDR
}

func (g *cidrListGeneric) Set(value string) error {
	m := g.value
	if *m == nil {
		n := make([]egoscale.CIDR, 0)
		*m = n
	}

	values := strings.Split(value, ",")

	for _, value := range values {
		cidr, err := egoscale.ParseCIDR(value)
		if err != nil {
			return err
		}
		*m = append(*m, *cidr)
	}
	return nil
}

func (g *cidrListGeneric) Type() string {
	return "cidrListGeneric"
}

func (g *cidrListGeneric) String() string {
	m := g.value
	if m == nil || *m == nil {
		return ""
	}
	vs := make([]string, 0, len(*m))
	for _, v := range *m {
		vs = append(vs, v.String())
	}

	return strings.Join(vs, ",")
}

// tag flag
type tagGeneric struct {
	value *[]egoscale.ResourceTag
}

func (g *tagGeneric) Set(value string) error {
	m := g.value
	if *m == nil {
		n := make([]egoscale.ResourceTag, 0)
		*m = n
	}

	keypairs := strings.Split(value, ",")
	for _, kv := range keypairs {
		values := strings.SplitN(kv, ":", 2)
		if len(values) != 2 {
			return fmt.Errorf("not a valid key:value content, got %s", kv)
		}

		*m = append(*m, egoscale.ResourceTag{
			Key:   values[0],
			Value: values[1],
		})
	}

	return nil
}

func (g *tagGeneric) Type() string {
	return "tag"
}

func (g *tagGeneric) String() string {
	m := g.value
	if m == nil || *m == nil {
		return ""
	}
	vs := make([]string, 0, len(*m))
	for _, v := range *m {
		vs = append(vs, fmt.Sprintf("%s=%s", v.Key, v.Value))
	}

	return strings.Join(vs, ",")
}

// map flag

type mapGeneric struct {
	value *map[string]string
}

func (g *mapGeneric) Set(value string) error {
	m := g.value
	if *m == nil {
		n := make(map[string]string)
		*m = n
	}

	values := strings.SplitN(value, "=", 2)
	if len(values) != 2 {
		return fmt.Errorf("not a valid key=value content, got %s", value)
	}

	(*m)[values[0]] = values[1]
	return nil
}

func (g *mapGeneric) Type() string {
	return "map"
}

func (g *mapGeneric) String() string {
	m := g.value
	if *m == nil {
		return ""
	}
	vs := make([]string, 0, len(*m))
	for k, v := range *m {
		vs = append(vs, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(vs, ",")
}

// stringer generic flag
type stringerTypeGeneric struct {
	addr  interface{}
	value string
	typ   reflect.Type
}

func (g *stringerTypeGeneric) Set(value string) error {
	tv := reflect.ValueOf(g.addr)
	fv := reflect.ValueOf(&value)

	tv.Elem().Set(fv.Convert(reflect.PtrTo(g.typ)).Elem())

	g.value = value

	return nil
}

func (g *stringerTypeGeneric) Type() string {
	return "string"
}

func (g *stringerTypeGeneric) String() string {
	return g.value
}

// int16 generic flag
type intTypeGeneric struct {
	addr    interface{}
	value   int64
	base    int
	bitSize int
	typ     reflect.Type
}

func (g *intTypeGeneric) Set(value string) error {
	v, err := strconv.ParseInt(value, g.base, g.bitSize)
	if err != nil {
		return err
	}

	tv := reflect.ValueOf(g.addr)
	var fv reflect.Value
	switch g.bitSize {
	case 8:
		val := (int8)(v)
		fv = reflect.ValueOf(&val)
	case 16:
		val := (int16)(v)
		fv = reflect.ValueOf(&val)
	case 32:
		val := (int)(v)
		fv = reflect.ValueOf(&val)
	case 64:
		fv = reflect.ValueOf(&v)
	}
	tv.Elem().Set(fv.Convert(reflect.PtrTo(g.typ)).Elem())

	g.value = v

	return nil
}

func (g *intTypeGeneric) Type() string {
	return "int16"
}

func (g *intTypeGeneric) String() string {
	if g.addr != nil {
		return strconv.FormatInt(g.value, g.base)
	}
	return ""
}
