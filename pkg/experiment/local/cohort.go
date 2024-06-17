package local

const userGroupType = "user"

type Cohort struct {
	ID           string
	LastModified int64
	Size         int
	MemberIDs    map[string]struct{}
	GroupType    string
}
