package evaluation

// IsCohortFilter checks if the condition is a cohort filter.
func (f Flag) IsCohortFilter(condition map[string]interface{}) bool {
	op, opExists := condition["op"].(string)
	selector, selectorExists := condition["selector"].([]interface{})
	if opExists && selectorExists && len(selector) > 0 && selector[len(selector)-1] == "cohort_ids" {
		return op == "set contains any" || op == "set does not contain any"
	}
	return false
}

// GetGroupedCohortConditionIDs extracts grouped cohort condition IDs from a segment.
func (f Flag) GetGroupedCohortConditionIDs(segment map[string]interface{}) map[string]map[string]bool {
	cohortIDs := make(map[string]map[string]bool)
	conditions, ok := segment["conditions"].([]interface{})
	if !ok {
		return cohortIDs
	}
	for _, outer := range conditions {
		outerCondition, ok := outer.(map[string]interface{})
		if !ok {
			continue
		}
		for _, condition := range outerCondition {
			conditionMap, ok := condition.(map[string]interface{})
			if !ok {
				continue
			}
			if f.IsCohortFilter(conditionMap) {
				selector, _ := conditionMap["selector"].([]interface{})
				if len(selector) > 2 {
					contextSubtype := selector[1].(string)
					var groupType string
					if contextSubtype == "user" {
						groupType = cohort.USER_GROUP_TYPE
					} else if selectorContainsGroups(selector) {
						groupType = selector[2].(string)
					} else {
						continue
					}
					values, _ := conditionMap["values"].([]interface{})
					cohortIDs[groupType] = map[string]bool{}
					for _, value := range values {
						cohortIDs[groupType][value.(string)] = true
					}
				}
			}
		}
	}
	return cohortIDs
}

// helper function to check if selector contains groups
func selectorContainsGroups(selector []interface{}) bool {
	for _, s := range selector {
		if s == "groups" {
			return true
		}
	}
	return false
}

// GetGroupedCohortIDsFromFlag extracts grouped cohort IDs from a flag.
func (f Flag) GetGroupedCohortIDsFromFlag(flag map[string]interface{}) map[string]map[string]bool {
	cohortIDs := make(map[string]map[string]bool)
	segments, ok := flag["segments"].([]interface{})
	if !ok {
		return cohortIDs
	}
	for _, seg := range segments {
		segment, _ := seg.(map[string]interface{})
		for key, values := range f.GetGroupedCohortConditionIDs(segment) {
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
func (f Flag) GetAllCohortIDsFromFlag(flag map[string]interface{}) map[string]bool {
	cohortIDs := make(map[string]bool)
	groupedIDs := f.GetGroupedCohortIDsFromFlag(flag)
	for _, values := range groupedIDs {
		for id := range values {
			cohortIDs[id] = true
		}
	}
	return cohortIDs
}

// GetAllCohortIDsFromFlags extracts all cohort IDs from multiple flags.
func (f Flag) GetAllCohortIDsFromFlags(flags []map[string]interface{}) map[string]bool {
	cohortIDs := make(map[string]bool)
	for _, flag := range flags {
		for id := range f.GetAllCohortIDsFromFlag(flag) {
			cohortIDs[id] = true
		}
	}
	return cohortIDs
}
