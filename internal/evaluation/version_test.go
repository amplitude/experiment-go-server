package evaluation

import (
	"testing"
)

func TestInvalidVersions(t *testing.T) {
	// just major
	assertInvalidVersion(t, "10")
	// trailing dots
	assertInvalidVersion(t, "10.")
	assertInvalidVersion(t, "10..")
	assertInvalidVersion(t, "10.2.")
	assertInvalidVersion(t, "10.2.33.")
	// trailing dots on prerelease tags are not handled because prerelease tags are considered
	// strings anyway for comparison which should be fine - e.g. "10.2.33-alpha1.2."

	// dots in the middle
	assertInvalidVersion(t, "10..2.33")
	assertInvalidVersion(t, "102...33")

	// invalid characters
	assertInvalidVersion(t, "a.2.3")
	assertInvalidVersion(t, "23!")
	assertInvalidVersion(t, "23.#5")
	assertInvalidVersion(t, "")

	// more numbers
	assertInvalidVersion(t, "2.3.4.567")
	assertInvalidVersion(t, "2.3.4.5.6.7")

	// prerelease if provided should always have major, minor, patch
	assertInvalidVersion(t, "10.2.alpha")
	assertInvalidVersion(t, "10.alpha")
	assertInvalidVersion(t, "alpha-1.2.3")

	// prerelease should be separated by a hyphen after patch
	assertInvalidVersion(t, "10.2.3alpha")
	assertInvalidVersion(t, "10.2.3alpha-1.2.3")

	// negative numbers
	assertInvalidVersion(t, "-10.1")
	assertInvalidVersion(t, "10.-1")
}

func TestValidVersions(t *testing.T) {
	assertValidVersions(t, "100.2")
	assertValidVersions(t, "0.102.39")
	assertValidVersions(t, "0.0.0")

	// versions with leading 0s would be converted to int
	assertValidVersions(t, "01.02")
	assertValidVersions(t, "001.001100.000900")

	// prerelease tags
	assertValidVersions(t, "10.20.30-alpha")
	assertValidVersions(t, "10.20.30-1.x.y")
	assertValidVersions(t, "10.20.30-aslkjd")
	assertValidVersions(t, "10.20.30-b894")
	assertValidVersions(t, "10.20.30-b8c9")
}

func TestVersionComparison(t *testing.T) {
	// EQUALS case
	assertVersionComparison(t, "66.12.23", OpIs, "66.12.23")
	// patch if not specified equals 0
	assertVersionComparison(t, "5.6", OpIs, "5.6.0")
	// leading 0s are not stored when parsed
	assertVersionComparison(t, "06.007.0008", OpIs, "6.7.8")
	// with pre release
	assertVersionComparison(t, "1.23.4-b-1.x.y", OpIs, "1.23.4-b-1.x.y")

	// DOES NOT EQUAL case
	assertVersionComparison(t, "1.23.4-alpha-1.2", OpIsNot, "1.23.4-alpha-1")
	// trailing 0s aren't stripped
	assertVersionComparison(t, "1.2.300", OpIsNot, "1.2.3")
	assertVersionComparison(t, "1.20.3", OpIsNot, "1.2.3")

	// LESS THAN case
	// patch of .1 makes it greater
	assertVersionComparison(t, "50.2", OpVersionLessThan, "50.2.1")
	// minor 9 > minor 20
	assertVersionComparison(t, "20.9", OpVersionLessThan, "20.20")
	// same version with pre release should be lesser
	assertVersionComparison(t, "20.9.4-alpha1", OpVersionLessThan, "20.9.4")
	// compare prerelease as strings
	assertVersionComparison(t, "20.9.4-a-1.2.3", OpVersionLessThan, "20.9.4-a-1.3")
	// since prerelease is compared as strings a1.23 < a1.5 because 2 < 5
	assertVersionComparison(t, "20.9.4-a1.23", OpVersionLessThan, "20.9.4-a1.5")

	// GREATER THAN case
	assertVersionComparison(t, "12.30.2", OpVersionGreaterThan, "12.4.1")
	// 100 > 1
	assertVersionComparison(t, "7.100", OpVersionGreaterThan, "7.1")
	// 10 > 9
	assertVersionComparison(t, "7.10", OpVersionGreaterThan, "7.9")
	// converts to 7.10.20 > 7.9.1
	assertVersionComparison(t, "07.010.0020", OpVersionGreaterThan, "7.009.1")
	// patch comparison comes first
	assertVersionComparison(t, "20.5.6-b1.2.x", OpVersionGreaterThan, "20.5.5")
}

func assertInvalidVersion(t *testing.T, ver string) {
	if parseVersion(ver) != nil {
		t.Fatalf("expected invalid version %v", ver)
	}
}

func assertValidVersions(t *testing.T, ver string) {
	if parseVersion(ver) == nil {
		t.Fatalf("expected valid version %v", ver)
	}
}

func assertVersionComparison(t *testing.T, v1, op, v2 string) {
	sv1 := parseVersion(v1)
	if sv1 == nil {
		t.Fatalf("expected valid version %v", v1)
	}
	sv2 := parseVersion(v2)
	if sv2 == nil {
		t.Fatalf("expected valid version %v", v2)
	}
	switch op {
	case OpIs:
		if versionCompare(*sv1, *sv2) != 0 {
			t.Fatalf("expected %v == %v", v1, v2)
		}
	case OpIsNot:
		if versionCompare(*sv1, *sv2) == 0 {
			t.Fatalf("expected %v != %v", v1, v2)
		}
	case OpVersionLessThan:
		if versionCompare(*sv1, *sv2) >= 0 {
			t.Fatalf("expected %v < %v", v1, v2)
		}
	case OpVersionGreaterThan:
		if versionCompare(*sv1, *sv2) <= 0 {
			t.Fatalf("expected %v < %v", v1, v2)
		}
	}
}
