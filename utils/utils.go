package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"

	v2 "github.com/exoscale/egoscale/v2"
)

// RandStringBytes Generate random string of n bytes
func RandStringBytes(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func GetInstancesInSecurityGroup(ctx context.Context, client *v2.Client, securityGroupID, zone string) ([]*v2.Instance, error) {
	allInstances, err := client.ListInstances(ctx, zone)
	if err != nil {
		return nil, err
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

func GetInstancesAttachedToEIP(ctx context.Context, client *v2.Client, elasticIPID, zone string) ([]*v2.Instance, error) {
	instances, err := client.ListInstances(ctx, zone, v2.ListInstancesByIpAddress(elasticIPID))
	if err != nil {
		return nil, err
	}

	return instances, nil
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
