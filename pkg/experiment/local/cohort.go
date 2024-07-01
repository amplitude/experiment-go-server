package local

import "sort"

const userGroupType = "User"

type Cohort struct {
	Id           string
	LastModified int64
	Size         int
	MemberIds    []string
	GroupType    string
}

func CohortEquals(c1, c2 *Cohort) bool {
	if c1.Id != c2.Id || c1.LastModified != c2.LastModified || c1.Size != c2.Size || c1.GroupType != c2.GroupType {
		return false
	}
	if len(c1.MemberIds) != len(c2.MemberIds) {
		return false
	}

	// Sort MemberIds before comparing
	sort.Strings(c1.MemberIds)
	sort.Strings(c2.MemberIds)

	for i := range c1.MemberIds {
		if c1.MemberIds[i] != c2.MemberIds[i] {
			return false
		}
	}

	return true
}
