package local

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/amplitude/experiment-go-server/internal/evaluation"
	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"github.com/amplitude/experiment-go-server/pkg/logger"
)

// recordingCohortStorage records which cohort IDs each membership lookup asks about.
type recordingCohortStorage struct {
	cohortStorage
	userLookups []map[string]struct{}
}

func (s *recordingCohortStorage) getCohortsForUser(userID string, cohortIDs map[string]struct{}) map[string]struct{} {
	s.userLookups = append(s.userLookups, cohortIDs)
	return s.cohortStorage.getCohortsForUser(userID, cohortIDs)
}

var enrichmentTestVariants = map[string]*evaluation.Variant{
	"on": {Key: "on", Value: "on"},
}

func userCohortFlag(key, cohortID string, dependencies ...string) *evaluation.Flag {
	return &evaluation.Flag{
		Key:          key,
		Variants:     enrichmentTestVariants,
		Dependencies: dependencies,
		Segments: []*evaluation.Segment{
			{
				Conditions: [][]*evaluation.Condition{
					{
						{
							Selector: []string{"context", "user", "cohort_ids"},
							Op:       "set contains any",
							Values:   []string{cohortID},
						},
					},
				},
				Variant: "on",
			},
		},
	}
}

func alwaysOnFlag(key string, dependencies ...string) *evaluation.Flag {
	return &evaluation.Flag{
		Key:          key,
		Variants:     enrichmentTestVariants,
		Dependencies: dependencies,
		Segments:     []*evaluation.Segment{{Variant: "on"}},
	}
}

// evaluateWithCohorts runs EvaluateV2WithOptions against in-memory flag and
// cohort storage where "user" belongs to cohort-a and cohort-b.
func evaluateWithCohorts(t *testing.T, flags []*evaluation.Flag, flagKeys []string) (map[string]experiment.Variant, *experiment.User, *recordingCohortStorage) {
	t.Helper()
	log := logger.New(logger.Error, logger.NewDefault())

	flagStorage := newInMemoryFlagConfigStorage()
	for _, flag := range flags {
		flagStorage.putFlagConfig(flag)
	}
	cohorts := &recordingCohortStorage{cohortStorage: newInMemoryCohortStorage()}
	cohorts.putCohort(&Cohort{Id: "cohort-a", Size: 1, MemberIds: []string{"user"}, GroupType: userGroupType})
	cohorts.putCohort(&Cohort{Id: "cohort-b", Size: 1, MemberIds: []string{"user"}, GroupType: userGroupType})

	client := &Client{
		log:               log,
		config:            &Config{},
		engine:            evaluation.NewEngine(log),
		cohortStorage:     cohorts,
		flagConfigStorage: flagStorage,
	}
	user := &experiment.User{UserId: "user"}
	variants, err := client.EvaluateV2WithOptions(user, &EvaluateOptions{FlagKeys: flagKeys})
	assert.NoError(t, err)
	return variants, user, cohorts
}

func TestEvaluateV2OnlyLooksUpCohortsOfRequestedFlags(t *testing.T) {
	flags := []*evaluation.Flag{
		userCohortFlag("flag-a", "cohort-a"),
		userCohortFlag("flag-b", "cohort-b"),
	}

	variants, user, cohorts := evaluateWithCohorts(t, flags, []string{"flag-a"})

	assert.Equal(t, "on", variants["flag-a"].Value)
	assert.NotContains(t, variants, "flag-b")
	assert.Equal(t, []map[string]struct{}{{"cohort-a": {}}}, cohorts.userLookups)
	assert.Equal(t, map[string]struct{}{"cohort-a": {}}, user.CohortIds)
}

func TestEvaluateV2LooksUpCohortsOfRequestedFlagDependencies(t *testing.T) {
	flags := []*evaluation.Flag{
		userCohortFlag("parent", "cohort-b"),
		alwaysOnFlag("child", "parent"),
		userCohortFlag("unrelated", "cohort-a"),
	}

	variants, user, cohorts := evaluateWithCohorts(t, flags, []string{"child"})

	assert.Equal(t, "on", variants["parent"].Value)
	assert.Equal(t, "on", variants["child"].Value)
	assert.Equal(t, []map[string]struct{}{{"cohort-b": {}}}, cohorts.userLookups)
	assert.Equal(t, map[string]struct{}{"cohort-b": {}}, user.CohortIds)
}

func TestEvaluateV2WithoutFlagKeysLooksUpAllCohorts(t *testing.T) {
	flags := []*evaluation.Flag{
		userCohortFlag("flag-a", "cohort-a"),
		userCohortFlag("flag-b", "cohort-b"),
	}

	variants, user, cohorts := evaluateWithCohorts(t, flags, nil)

	assert.Equal(t, "on", variants["flag-a"].Value)
	assert.Equal(t, "on", variants["flag-b"].Value)
	assert.Equal(t, []map[string]struct{}{{"cohort-a": {}, "cohort-b": {}}}, cohorts.userLookups)
	assert.Equal(t, map[string]struct{}{"cohort-a": {}, "cohort-b": {}}, user.CohortIds)
}

func TestEvaluateV2SkipsCohortLookupWhenRequestedFlagsTargetNoCohorts(t *testing.T) {
	flags := []*evaluation.Flag{
		alwaysOnFlag("plain"),
		userCohortFlag("flag-a", "cohort-a"),
	}

	variants, user, cohorts := evaluateWithCohorts(t, flags, []string{"plain"})

	assert.Equal(t, "on", variants["plain"].Value)
	assert.Empty(t, cohorts.userLookups)
	assert.Nil(t, user.CohortIds)
}
