package local

import "hash/fnv"

func hashCode(s string) uint64 {
	// Create a new FNV-1a hash
	hash := fnv.New64a()

	// Write the bytes of the string to the hash
	hash.Write([]byte(s))

	// Return the resulting hash code as a 64-bit unsigned integer
	return hash.Sum64()
}
