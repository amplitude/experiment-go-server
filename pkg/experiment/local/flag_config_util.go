package local

import (
	"github.com/amplitude/experiment-go-server/internal/evaluation"
)

// isCohortFilter checks if the condition is a cohort filter.
func isCohortFilter(condition *evaluation.Condition) bool {
	op := condition.Op
	selector := condition.Selector
	if len(selector) > 0 && selector[len(selector)-1] == "cohort_ids" {
		return op == "set contains any" || op == "set does not contain any"
	}
	return false
}

// getGroupedCohortConditionIDs extracts grouped cohort condition IDs from a segment.
func getGroupedCohortConditionIDs(segment *evaluation.Segment) map[string]map[string]struct{} {
	cohortIDs := make(map[string]map[string]struct{})
	if segment == nil {
		return cohortIDs
	}

	for _, outer := range segment.Conditions {
		for _, condition := range outer {
			if isCohortFilter(condition) {
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
					cohortIDs[groupType] = make(map[string]struct{})
					for _, value := range values {
						cohortIDs[groupType][value] = struct{}{}
					}
				}
			}
		}
	}
	return cohortIDs
}

// getGroupedCohortIDsFromFlag extracts grouped cohort IDs from a flag.
func getGroupedCohortIDsFromFlag(flag *evaluation.Flag) map[string]map[string]struct{} {
	cohortIDs := make(map[string]map[string]struct{})
	for _, segment := range flag.Segments {
		for key, values := range getGroupedCohortConditionIDs(segment) {
			if _, exists := cohortIDs[key]; !exists {
				cohortIDs[key] = make(map[string]struct{})
			}
			for id := range values {
				cohortIDs[key][id] = struct{}{}
			}
		}
	}
	return cohortIDs
}

// getAllCohortIDsFromFlag extracts all cohort IDs from a flag.
func getAllCohortIDsFromFlag(flag *evaluation.Flag) map[string]struct{} {
	cohortIDs := make(map[string]struct{})
	groupedIDs := getGroupedCohortIDsFromFlag(flag)
	for _, values := range groupedIDs {
		for id := range values {
			cohortIDs[id] = struct{}{}
		}
	}
	return cohortIDs
}

// getGroupedCohortIDsFromFlags extracts grouped cohort IDs from multiple flags.
func getGroupedCohortIDsFromFlags(flags []*evaluation.Flag) map[string]map[string]struct{} {
	cohortIDs := make(map[string]map[string]struct{})
	for _, flag := range flags {
		for key, values := range getGroupedCohortIDsFromFlag(flag) {
			if _, exists := cohortIDs[key]; !exists {
				cohortIDs[key] = make(map[string]struct{})
			}
			for id := range values {
				cohortIDs[key][id] = struct{}{}
			}
		}
	}
	return cohortIDs
}

// getAllCohortIDsFromFlags extracts all cohort IDs from multiple flags.
func getAllCohortIDsFromFlags(flags []*evaluation.Flag) map[string]struct{} {
	cohortIDs := make(map[string]struct{})
	for _, flag := range flags {
		for id := range getAllCohortIDsFromFlag(flag) {
			cohortIDs[id] = struct{}{}
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
