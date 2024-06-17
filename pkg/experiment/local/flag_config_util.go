package local

import (
	"github.com/amplitude/experiment-go-server/internal/evaluation"
)

// IsCohortFilter checks if the condition is a cohort filter.
func IsCohortFilter(condition *evaluation.Condition) bool {
	op := condition.Op
	selector := condition.Selector
	if len(selector) > 0 && selector[len(selector)-1] == "cohort_ids" {
		return op == "set contains any" || op == "set does not contain any"
	}
	return false
}

// GetGroupedCohortConditionIDs extracts grouped cohort condition IDs from a segment.
func GetGroupedCohortConditionIDs(segment *evaluation.Segment) map[string]map[string]bool {
	cohortIDs := make(map[string]map[string]bool)
	if segment == nil {
		return cohortIDs
	}

	for _, outer := range segment.Conditions {
		for _, condition := range outer {
			if IsCohortFilter(condition) {
				selector := condition.Selector
				if len(selector) > 2 {
					contextSubtype := selector[1]
					var groupType string
					if contextSubtype == "user" {
						groupType = userGroupType
					} else if selectorContainsGroups(selector) {
						groupType = selector[2]
					} else {
						continue
					}
					values := condition.Values
					cohortIDs[groupType] = make(map[string]bool)
					for _, value := range values {
						cohortIDs[groupType][value] = true
					}
				}
			}
		}
	}
	return cohortIDs
}

// GetGroupedCohortIDsFromFlag extracts grouped cohort IDs from a flag.
func GetGroupedCohortIDsFromFlag(flag *evaluation.Flag) map[string]map[string]bool {
	cohortIDs := make(map[string]map[string]bool)
	for _, segment := range flag.Segments {
		for key, values := range GetGroupedCohortConditionIDs(segment) {
			if _, exists := cohortIDs[key]; !exists {
				cohortIDs[key] = make(map[string]bool)
			}
			for id := range values {
				cohortIDs[key][id] = true
			}
		}
	}
	return cohortIDs
}

// GetAllCohortIDsFromFlag extracts all cohort IDs from a flag.
func GetAllCohortIDsFromFlag(flag *evaluation.Flag) map[string]bool {
	cohortIDs := make(map[string]bool)
	groupedIDs := GetGroupedCohortIDsFromFlag(flag)
	for _, values := range groupedIDs {
		for id := range values {
			cohortIDs[id] = true
		}
	}
	return cohortIDs
}

// GetGroupedCohortIDsFromFlags extracts grouped cohort IDs from multiple flags.
func GetGroupedCohortIDsFromFlags(flags []*evaluation.Flag) map[string]map[string]bool {
	cohortIDs := make(map[string]map[string]bool)
	for _, flag := range flags {
		for key, values := range GetGroupedCohortIDsFromFlag(flag) {
			if _, exists := cohortIDs[key]; !exists {
				cohortIDs[key] = make(map[string]bool)
			}
			for id := range values {
				cohortIDs[key][id] = true
			}
		}
	}
	return cohortIDs
}

// GetAllCohortIDsFromFlags extracts all cohort IDs from multiple flags.
func GetAllCohortIDsFromFlags(flags []*evaluation.Flag) map[string]bool {
	cohortIDs := make(map[string]bool)
	for _, flag := range flags {
		for id := range GetAllCohortIDsFromFlag(flag) {
			cohortIDs[id] = true
		}
	}
	return cohortIDs
}

// helper function to check if selector contains groups
func selectorContainsGroups(selector []string) bool {
	for _, s := range selector {
		if s == "groups" {
			return true
		}
	}
	return false
}
