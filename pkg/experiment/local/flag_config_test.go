package local

import (
	"testing"

	"github.com/amplitude/experiment-go-server/internal/evaluation"
	"github.com/stretchr/testify/assert"
)

func TestGetAllCohortIDsFromFlag(t *testing.T) {
	flags := getTestFlags()
	expectedCohortIDs := []string{
		"cohort1", "cohort2", "cohort3", "cohort4", "cohort5", "cohort6", "cohort7", "cohort8",
	}
	expectedCohortIDSet := make(map[string]bool)
	for _, id := range expectedCohortIDs {
		expectedCohortIDSet[id] = true
	}

	for _, flag := range flags {
		cohortIDs := getAllCohortIDsFromFlag(flag)
		for id := range cohortIDs {
			assert.True(t, expectedCohortIDSet[id])
		}
	}
}

func TestGetGroupedCohortIDsFromFlag(t *testing.T) {
	flags := getTestFlags()
	expectedGroupedCohortIDs := map[string][]string{
		"User":       {"cohort1", "cohort2", "cohort3", "cohort4", "cohort5", "cohort6"},
		"group_name": {"cohort7", "cohort8"},
	}

	for _, flag := range flags {
		groupedCohortIDs := getGroupedCohortIDsFromFlag(flag)
		for key, values := range groupedCohortIDs {
			assert.Contains(t, expectedGroupedCohortIDs, key)
			expectedSet := make(map[string]bool)
			for _, id := range expectedGroupedCohortIDs[key] {
				expectedSet[id] = true
			}
			for id := range values {
				assert.True(t, expectedSet[id])
			}
		}
	}
}

func TestGetAllCohortIDsFromFlags(t *testing.T) {
	flags := getTestFlags()
	expectedCohortIDs := []string{
		"cohort1", "cohort2", "cohort3", "cohort4", "cohort5", "cohort6", "cohort7", "cohort8",
	}
	expectedCohortIDSet := make(map[string]bool)
	for _, id := range expectedCohortIDs {
		expectedCohortIDSet[id] = true
	}

	cohortIDs := getAllCohortIDsFromFlags(flags)
	for id := range cohortIDs {
		assert.True(t, expectedCohortIDSet[id])
	}
}

func TestGetGroupedCohortIDsFromFlags(t *testing.T) {
	flags := getTestFlags()
	expectedGroupedCohortIDs := map[string][]string{
		"User":       {"cohort1", "cohort2", "cohort3", "cohort4", "cohort5", "cohort6"},
		"group_name": {"cohort7", "cohort8"},
	}

	groupedCohortIDs := getGroupedCohortIDsFromFlags(flags)
	for key, values := range groupedCohortIDs {
		assert.Contains(t, expectedGroupedCohortIDs, key)
		expectedSet := make(map[string]bool)
		for _, id := range expectedGroupedCohortIDs[key] {
			expectedSet[id] = true
		}
		for id := range values {
			assert.True(t, expectedSet[id])
		}
	}
}

func getTestFlags() []*evaluation.Flag {
	return []*evaluation.Flag{
		{
			Key: "flag-1",
			Metadata: map[string]interface{}{
				"deployed":       true,
				"evaluationMode": "local",
				"flagType":       "release",
				"flagVersion":    1,
			},
			Segments: []*evaluation.Segment{
				{
					Conditions: [][]*evaluation.Condition{
						{
							{
								Op:       "set contains any",
								Selector: []string{"context", "user", "cohort_ids"},
								Values:   []string{"cohort1", "cohort2"},
							},
						},
					},
					Metadata: map[string]interface{}{
						"segmentName": "Segment A",
					},
					Variant: "on",
				},
				{
					Metadata: map[string]interface{}{
						"segmentName": "All Other Users",
					},
					Variant: "off",
				},
			},
			Variants: map[string]*evaluation.Variant{
				"off": {
					Key: "off",
					Metadata: map[string]interface{}{
						"default": true,
					},
				},
				"on": {
					Key:   "on",
					Value: "on",
				},
			},
		},
		{
			Key: "flag-2",
			Metadata: map[string]interface{}{
				"deployed":       true,
				"evaluationMode": "local",
				"flagType":       "release",
				"flagVersion":    2,
			},
			Segments: []*evaluation.Segment{
				{
					Conditions: [][]*evaluation.Condition{
						{
							{
								Op:       "set contains any",
								Selector: []string{"context", "user", "cohort_ids"},
								Values:   []string{"cohort3", "cohort4", "cohort5", "cohort6"},
							},
						},
					},
					Metadata: map[string]interface{}{
						"segmentName": "Segment B",
					},
					Variant: "on",
				},
				{
					Metadata: map[string]interface{}{
						"segmentName": "All Other Users",
					},
					Variant: "off",
				},
			},
			Variants: map[string]*evaluation.Variant{
				"off": {
					Key: "off",
					Metadata: map[string]interface{}{
						"default": true,
					},
				},
				"on": {
					Key:   "on",
					Value: "on",
				},
			},
		},
		{
			Key: "flag-3",
			Metadata: map[string]interface{}{
				"deployed":       true,
				"evaluationMode": "local",
				"flagType":       "release",
				"flagVersion":    3,
			},
			Segments: []*evaluation.Segment{
				{
					Conditions: [][]*evaluation.Condition{
						{
							{
								Op:       "set contains any",
								Selector: []string{"context", "groups", "group_name", "cohort_ids"},
								Values:   []string{"cohort7", "cohort8"},
							},
						},
					},
					Metadata: map[string]interface{}{
						"segmentName": "Segment C",
					},
					Variant: "on",
				},
				{
					Metadata: map[string]interface{}{
						"segmentName": "All Other Groups",
					},
					Variant: "off",
				},
			},
			Variants: map[string]*evaluation.Variant{
				"off": {
					Key: "off",
					Metadata: map[string]interface{}{
						"default": true,
					},
				},
				"on": {
					Key:   "on",
					Value: "on",
				},
			},
		},
	}
}
