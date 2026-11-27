package local

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCohortsForUserWithUnsortedMembers(t *testing.T) {
	storage := newInMemoryCohortStorage()
	// The cohort API does not guarantee member order.
	storage.putCohort(&Cohort{Id: "a", Size: 4, MemberIds: []string{"zed", "1", "bob", "10"}, GroupType: userGroupType})
	storage.putCohort(&Cohort{Id: "b", Size: 2, MemberIds: []string{"bob", "amy"}, GroupType: userGroupType})

	requested := map[string]struct{}{"a": {}, "b": {}}

	for _, member := range []string{"zed", "1", "10"} {
		assert.Equal(t, map[string]struct{}{"a": {}}, storage.getCohortsForUser(member, requested), member)
	}
	assert.Equal(t, map[string]struct{}{"a": {}, "b": {}}, storage.getCohortsForUser("bob", requested))
	assert.Equal(t, map[string]struct{}{"b": {}}, storage.getCohortsForUser("amy", requested))
	assert.Empty(t, storage.getCohortsForUser("2", requested))
	assert.Empty(t, storage.getCohortsForUser("", requested))

	// Only requested cohorts are returned.
	assert.Equal(t, map[string]struct{}{"b": {}}, storage.getCohortsForUser("bob", map[string]struct{}{"b": {}}))
}

func TestGetCohortsForGroupIsScopedToGroupType(t *testing.T) {
	storage := newInMemoryCohortStorage()
	storage.putCohort(&Cohort{Id: "org", Size: 2, MemberIds: []string{"Amplitude", "Acme"}, GroupType: "org name"})
	storage.putCohort(&Cohort{Id: "users", Size: 1, MemberIds: []string{"Amplitude"}, GroupType: userGroupType})

	requested := map[string]struct{}{"org": {}, "users": {}}

	assert.Equal(t, map[string]struct{}{"org": {}}, storage.getCohortsForGroup("org name", "Amplitude", requested))
	assert.Equal(t, map[string]struct{}{"org": {}}, storage.getCohortsForGroup("org name", "Acme", requested))
	assert.Empty(t, storage.getCohortsForGroup("org name", "Other", requested))
	assert.Empty(t, storage.getCohortsForGroup("unknown type", "Amplitude", requested))
}

func TestGetCohortsForUserWithEmptyCohort(t *testing.T) {
	storage := newInMemoryCohortStorage()
	storage.putCohort(&Cohort{Id: "empty", Size: 0, MemberIds: []string{}, GroupType: userGroupType})
	storage.putCohort(&Cohort{Id: "nil", Size: 0, MemberIds: nil, GroupType: userGroupType})

	assert.Empty(t, storage.getCohortsForUser("1", map[string]struct{}{"empty": {}, "nil": {}}))
}

func BenchmarkGetCohortsForUser(b *testing.B) {
	for _, size := range []int{1_000, 1_000_000} {
		b.Run(fmt.Sprintf("members=%d", size), func(b *testing.B) {
			members := make([]string, size)
			for i := range members {
				members[i] = fmt.Sprintf("user-%d", size-i) // reverse order
			}
			storage := newInMemoryCohortStorage()
			storage.putCohort(&Cohort{Id: "a", Size: size, MemberIds: members, GroupType: userGroupType})
			requested := map[string]struct{}{"a": {}}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				storage.getCohortsForUser("missing-user", requested)
			}
		})
	}
}
