package local

func hashCode(s string) int {
	hash := 0
	if len(s) == 0 {
		return hash
	}
	for i := 0; i < len(s); i++ {
		chr := int(s[i])
		hash = (hash << 5) - hash + chr
		hash &= 0xFFFFFFFF
	}
	return hash
}

func difference(set1, set2 map[string]struct{}) map[string]struct{} {
	diff := make(map[string]struct{})
	for k := range set1 {
		if _, exists := set2[k]; !exists {
			diff[k] = struct{}{}
		}
	}
	return diff
}
