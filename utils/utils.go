package utils

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"

	exoapi "github.com/exoscale/egoscale/v2/api"
	v3 "github.com/exoscale/egoscale/v3"

	"github.com/exoscale/cli/pkg/account"
	v2 "github.com/exoscale/egoscale/v2"
	"github.com/exoscale/egoscale/v2/oapi"
)

var (
	// utils.AllZones represents the list of known Exoscale zones, in case we need it without performing API lookup.
	AllZones = []string{
		string(oapi.ZoneNameAtVie1),
		string(oapi.ZoneNameAtVie2),
		string(oapi.ZoneNameBgSof1),
		string(oapi.ZoneNameChDk2),
		string(oapi.ZoneNameChGva2),
		string(oapi.ZoneNameDeFra1),
		string(oapi.ZoneNameDeMuc1),
		string(oapi.ZoneNameHrZag1),
	}
)

func AllZonesV3(ctx context.Context, client v3.Client) ([]v3.ZoneName, error) {
	zones, err := client.ListZones(ctx)
	if err != nil {
		return nil, err
	}

	zoneNames := make([]v3.ZoneName, len(zones.Zones))

	for i, z := range zones.Zones {
		zoneNames[i] = z.Name
	}

	return zoneNames, nil
}

// RandStringBytes Generate random string of n bytes
func RandStringBytes(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func GetInstancesInSecurityGroup(ctx context.Context, client *v2.Client, securityGroupID string) ([]*v2.Instance, error) {
	allInstances := make([]*v2.Instance, 0)
	err := ForEachZone(AllZones, func(zone string) error {
		ctx := exoapi.WithEndpoint(ctx, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

		instances, err := client.ListInstances(ctx, zone)
		if err != nil {
			if !errors.Is(err, exoapi.ErrNotFound) {
				return err
			}
		} else {
			allInstances = append(allInstances, instances...)
		}

		return nil
	})
	if err != nil {
		if allInstances == nil {
			return nil, err
		} else {
			fmt.Printf("error while listing instances in security group: %s", err)
		}
	}

	var instancesInSG []*v2.Instance
	for _, instance := range allInstances {
		if instance.SecurityGroupIDs == nil {
			continue
		}

		for _, sgID := range *instance.SecurityGroupIDs {
			if sgID == securityGroupID {
				instancesInSG = append(instancesInSG, instance)
			}
		}
	}

	return instancesInSG, nil
}

func GetInstancesAttachedToEIP(ctx context.Context, client *v3.Client, elasticIPID string) ([]v3.ListInstancesResponseInstances, error) {
	instanceListResponse, err := client.ListInstances(ctx, v3.ListInstancesWithIPAddress(elasticIPID))
	if err != nil {
		return nil, err
	}

	return instanceListResponse.Instances, nil
}

// IsInList returns true if v exists in the specified list, false otherwise.
func IsInList(list []string, v string) bool {
	for _, lv := range list {
		if lv == v {
			return true
		}
	}

	return false
}

// EllipString truncates the string s with an ellipsis character if longer
// than maxLen.
func EllipString(s string, maxLen int) string {
	ellipsis := "…"

	if len(s) > maxLen {
		return s[0:maxLen-1] + ellipsis
	}

	return s
}

// DefaultString returns the value of the string pointer s if not nil, otherwise the default value specified.
func DefaultString(s *string, def string) string {
	if s != nil {
		return *s
	}

	return def
}

// DefaultBool returns the value of the bool pointer b if not nil, otherwise the default value specified.
func DefaultBool(b *bool, def bool) bool {
	if b != nil {
		return *b
	}

	return def
}

// DefaultIP returns the IP as string if not nil, otherwise the default value specified.
func DefaultIP(i *net.IP, def string) string {
	if i != nil {
		return i.String()
	}

	return def
}

// DefaultInt64 returns the value of the int64 pointer b if not nil, otherwise the default value specified.
func DefaultInt64(i *int64, def int64) int64 {
	if i != nil {
		return *i
	}

	return def
}

// NonEmptyStringPtr returns a non-nil pointer to s if the string is not empty, otherwise nil.
func NonEmptyStringPtr(s string) *string {
	if s != "" {
		return &s
	}

	return nil
}

func IsEmptyStringPtr(s *string) bool {
	return s == nil || *s == ""
}

// SliceToMap returns a map[string]string from a slice of KEY=VALUE formatted
// strings.
// This function is used to obtain a map[string]string from CLI flags, as the
// current CLI flags parsing module used (github.com/spf13/pflag) implements
// a "StringToString" type flag but doesn't support passing empty values,
// which we need in some cases (e.g. resetting labels).
func SliceToMap(v []string) (map[string]string, error) {
	m := make(map[string]string)

	for i := range v {
		parts := strings.SplitN(v[i], "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid value %q, expected format KEY=VALUE", v[i])
		}

		m[parts[0]] = parts[1]
	}

	return m, nil
}

// VersionMajor returns major part of a version number (given "x.y(.z)", returns "x").
// If the input version is not in semver format, returns 0.
func VersionMajor(version string) uint32 {
	parts := strings.Split(version, ".")

	if len(parts) > 0 {
		v, e := strconv.ParseUint(parts[0], 10, 32)
		if e != nil {
			return 0
		}

		return uint32(v)
	}

	return 0
}

// VersionMinor returns minor part of a version number (given "x.y(.z)", returns "y").
// If the input version is not in semver format, returns 0.
func VersionMinor(version string) uint32 {
	parts := strings.Split(version, ".")

	if len(parts) > 1 {
		v, e := strconv.ParseUint(parts[1], 10, 32)
		if e != nil {
			return 0
		}

		return uint32(v)
	}

	return 0
}

// VersionIsNewer returns true if new version has potential deprecation
func VersionIsNewer(old, new string) bool {
	return (VersionMajor(new) >= VersionMajor(old)) ||
		(VersionMajor(new) == VersionMajor(old) && VersionMinor(new) >= VersionMinor(old))
}

// VersionsAreEquivalent returns true if new and old versions both have same major and minor numbers
func VersionsAreEquivalent(a, b string) bool {
	return (VersionMajor(b) == VersionMajor(a) && VersionMinor(b) == VersionMinor(a))
}

// ForEachZone executes the function f for each specified zone, and return a multierror.Error containing all
// errors that may have occurred during execution.
func ForEachZone[T any](zones []T, f func(zone T) error) error {
	meg := new(multierror.Group)

	for _, zone := range zones {
		zone := zone
		meg.Go(func() error {
			return f(zone)
		})
	}

	return meg.Wait().ErrorOrNil()
}

// ForEveryZone executes the function f for every specified zone, and returns a multierror.Error containing all
// errors that may have occurred during execution.

// TODO: This is a copy paste from the function above, but suitable for egoscale v3 calls.
// Remove the old one after the migration.
func ForEveryZone(zones []v3.Zone, f func(zone v3.Zone) error) error {
	meg := new(multierror.Group)

	for _, zone := range zones {
		zone := zone
		meg.Go(func() error {
			return f(zone)
		})
	}

	return meg.Wait().ErrorOrNil()
}

// ParseInstanceType returns an v3.InstanceType with family and name.
func ParseInstanceType(instanceType string) v3.InstanceType {
	var typeFamily, typeSize string

	parts := strings.SplitN(instanceType, ".", 2)
	if l := len(parts); l > 0 {
		if l == 1 {
			typeFamily, typeSize = "standard", strings.ToLower(parts[0])
		} else {
			typeFamily, typeSize = strings.ToLower(parts[0]), strings.ToLower(parts[1])
		}
	}

	return v3.InstanceType{
		Family: v3.InstanceTypeFamily(typeFamily),
		Size:   v3.InstanceTypeSize(typeSize),
	}
}

// GetSettingFloat64 safely retrieves a float64 value from settings map and converts to int
func GetSettingFloat64(settings map[string]interface{}, key string) int {
	if val, ok := settings[key]; ok && val != nil {
		if fVal, ok := val.(float64); ok {
			return int(fVal)
		}
	}
	return 0
}

// GetSettingString safely retrieves a string value from settings map
func GetSettingString(settings map[string]interface{}, key string) string {
	if val, ok := settings[key]; ok && val != nil {
		if sVal, ok := val.(string); ok {
			return sVal
		}
	}
	return ""
}

// GetSettingBool safely retrieves a bool value from settings map
func GetSettingBool(settings map[string]interface{}, key string) bool {
	if val, ok := settings[key]; ok && val != nil {
		if bVal, ok := val.(bool); ok {
			return bVal
		}
	}
	return false
}

func ReadInput(ctx context.Context, reader *bufio.Reader, text, def string) (string, error) {
	if def == "" {
		fmt.Printf("[+] %s [%s]: ", text, "none")
	} else {
		fmt.Printf("[+] %s [%s]: ", text, def)
	}
	c := make(chan bool)
	defer close(c)

	input := ""
	var err error
	go func() {
		input, err = reader.ReadString('\n')
		c <- true
	}()

	select {
	case <-c:
	case <-ctx.Done():
		err = fmt.Errorf("")
	}

	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		input = def
	}
	return input, nil
}

func AskQuestion(ctx context.Context, text string) bool {
	reader := bufio.NewReader(os.Stdin)

	resp, err := ReadInput(ctx, reader, text, "yN")
	if err != nil {
		log.Fatal(err)
	}

	return (strings.ToLower(resp) == "y" || strings.ToLower(resp) == "yes")
}
