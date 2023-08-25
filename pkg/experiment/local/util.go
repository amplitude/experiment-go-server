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
