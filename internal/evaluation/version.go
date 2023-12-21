package evaluation

import (
	"regexp"
	"strconv"
)

type version struct  {
	major int
	minor int
	patch int
	preRelease string
}

const versionPattern = "^(\\d+)\\.(\\d+)(\\.(\\d+)(-(([-\\w]+\\.?)*))?)?$"
var regex, _ = regexp.Compile(versionPattern)

func parseVersion(versionString string) *version {
	if len(versionString) == 0 {
		return nil
	}
	matchGroup := regex.FindStringSubmatch(versionString)
	if matchGroup == nil {
		return nil
	}
	major, err := strconv.Atoi(matchGroup[1])
	if err != nil {
		return nil
	}
	minor, err := strconv.Atoi(matchGroup[2])
	if err != nil {
		return nil
	}
	patch, _ := strconv.Atoi(matchGroup[4])
	preRelease := matchGroup[5]
	return &version{major, minor, patch, preRelease}
}

func versionCompare(v1, v2 version) int {
	switch {
	case v1.major > v2.major: return 1
	case v1.major < v2.major: return -1
	case v1.minor > v2.minor: return 1
	case v1.minor < v2.minor: return -1
	case v1.patch > v2.patch: return 1
	case v1.patch < v2.patch: return -1
	case len(v1.preRelease) > 0 && len(v2.preRelease) == 0: return -1
	case len(v1.preRelease) == 0 && len(v2.preRelease) > 0: return 1
	case len(v1.preRelease) > 0 && len(v2.preRelease) > 0:
		return stringCompare(v1.preRelease, v2.preRelease)
	default: return 0
	}
}

func stringCompare(s1, s2 string) int {
	switch {
	case s1 < s2: return -1
	case s1 > s2: return 1
	default: return 0
	}
}
