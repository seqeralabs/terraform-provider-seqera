package stringvalidators

import "sort"

// mapKeys returns the keys of a map[string][]string as a sorted string slice.
// Used by CredentialsConfigValidator and ComputeConfigValidator for error messages.
func mapKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
