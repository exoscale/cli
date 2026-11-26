package instance

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// instanceListLabelsFilter formats labels for the labels query parameter:
// "k1=v1 k2=v2". Labels the API can't parse (empty, or containing a space
// or "=") are refused, as the API would silently ignore them.
func instanceListLabelsFilter(labels map[string]string) (string, error) {
	pairs := make([]string, 0, len(labels))
	for _, k := range slices.Sorted(maps.Keys(labels)) {
		v := labels[k]
		if k == "" || v == "" || strings.ContainsAny(k, " =") || strings.ContainsAny(v, " =") {
			return "", fmt.Errorf(
				"invalid label %q: key and value must be non-empty and must not contain spaces or '='",
				k+"="+v)
		}
		pairs = append(pairs, k+"="+v)
	}

	return strings.Join(pairs, " "), nil
}
