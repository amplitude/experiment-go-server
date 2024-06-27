package local

import "sort"

const userGroupType = "user"

type Cohort struct {
	ID           string
	LastModified int64
	Size         int
	MemberIDs    []string
	GroupType    string
}

func CohortEquals(c1, c2 *Cohort) bool {
	if c1.ID != c2.ID || c1.LastModified != c2.LastModified || c1.Size != c2.Size || c1.GroupType != c2.GroupType {
		return false
	}
	if len(c1.MemberIDs) != len(c2.MemberIDs) {
		return false
	}

	// Sort MemberIDs before comparing
	sort.Strings(c1.MemberIDs)
	sort.Strings(c2.MemberIDs)

	for i := range c1.MemberIDs {
		if c1.MemberIDs[i] != c2.MemberIDs[i] {
			return false
		}
	}

	return true
}
