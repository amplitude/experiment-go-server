package evaluation

import (
	"encoding/json"
	"fmt"
	"github.com/amplitude/experiment-go-server/pkg/logger"
	"github.com/spaolacci/murmur3"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Engine struct {
	log *logger.Logger
}

type target struct {
	context map[string]interface{}
	result  map[string]Variant
}

func NewEngine(log *logger.Logger) *Engine {
	return &Engine{log}
}

func (e *Engine) Evaluate(context map[string]interface{}, flags []*Flag) map[string]Variant {
	e.log.Debug("Evaluating %v flags with context %v", len(flags), context)
	results := make(map[string]Variant)
	target := &target{context, results}
	for _, flag := range flags {
		// Evaluate flag and update results
		variant := e.evaluateFlag(target, flag)
		if variant != nil {
			results[flag.Key] = *variant
		} else {
			e.log.Debug("Flag %v evaluation returned nil result", flag.Key)
		}
	}
	e.log.Debug("Evaluation completed. %v", results)
	return results
}

func (e *Engine) evaluateFlag(target *target, flag *Flag) *Variant {
	e.log.Verbose("Evaluating flag %v with target %v", flag, target)
	var result *Variant
	for _, segment := range flag.Segments {
		result = e.evaluateSegment(target, flag, segment)
		if result != nil {
			// Merge all metadata into the result
			metadata := mergeMetadata([]map[string]interface{}{flag.Metadata, segment.Metadata, result.Metadata})
			result = &Variant{result.Key, result.Value, result.Payload, metadata}
			e.log.Verbose("Flag evaluation returned result %v on segment %v", result, segment)
			break
		}
	}
	return result
}

func (e *Engine) evaluateSegment(target *target, flag *Flag, segment *Segment) *Variant {
	e.log.Verbose("Evaluating segment %v with target %")
	if segment.Conditions == nil {
		e.log.Verbose("Segment conditions are nil, bucketing target")
		// Null conditions always match
		variantKey := e.bucket(target, segment)
		return flag.Variants[variantKey]
	}
	// Outer list logic is "or" (||)
	for _, conditions := range segment.Conditions {
		match := true
		// Inner list logic is "and" (&&)
		for _, condition := range conditions {
			match = e.matchCondition(target, condition)
			if !match {
				e.log.Verbose("Segment condition %v did not match target", condition)
				break
			} else {
				e.log.Verbose("Segment condition %v matched target", condition)
			}
		}
		// On match bucket the user
		if match {
			e.log.Verbose("Segment conditions matched, bucketing target")
			variantKey := e.bucket(target, segment)
			return flag.Variants[variantKey]
		}
	}
	return nil
}

func (e *Engine) matchCondition(target *target, condition *Condition) bool {
	propValue := selectEach(target, condition.Selector)
	// We need special matching for null properties and set type prop values
	// and operators. All other values are matched as strings, since the
	// filter values are always strings.
	if propValue == nil {
		return matchNull(condition.Op, condition.Values)
	} else if isSetOperator(condition.Op) {
		propValueStringList, err := coerceStringList(propValue)
		if err != nil {
			return false
		}
		return matchSet(propValueStringList, condition.Op, condition.Values)
	} else {
		propValueString := coerceString(propValue)
		if propValueString == nil {
			return false
		}
		return matchString(*propValueString, condition.Op, condition.Values)
	}
}

func (e *Engine) getHash(key string) uint64 {
	return uint64(murmur3.Sum32WithSeed([]byte(key), 0))
}

func (e *Engine) bucket(target *target, segment *Segment) string {
	e.log.Verbose("Bucketing segment %v with target %v", segment, target)
	if segment.Bucket == nil {
		// A nil bucket means the segment is fully rolled out. Select the default variant.
		e.log.Verbose("Segment bucket is nil, returning default variant %v", segment.Variant)
		return segment.Variant
	}
	// Select the bucketing value
	bucketingValue := coerceString(selectEach(target, segment.Bucket.Selector))
	e.log.Verbose("Selected bucketing value %v from target", bucketingValue)
	if bucketingValue == nil || len(*bucketingValue) == 0 {
		// A nil or empty bucketing value cannot be bucketed. Select the default variant.
		e.log.Verbose("Selected bucketing value is nil or empty")
		return segment.Variant
	}
	// Salt and hash the value, and compute the allocation and distribution values.
	keyToHash := fmt.Sprintf("%v/%v", segment.Bucket.Salt, *bucketingValue)
	hash := e.getHash(keyToHash)
	allocationValue := hash % 100
	distributionValue := hash / 100
	for _, allocation := range segment.Bucket.Allocations {
		allocationStart := allocation.Range[0]
		allocationEnd := allocation.Range[1]
		if allocationValue >= allocationStart && allocationValue < allocationEnd {
			for _, distribution := range allocation.Distributions {
				distributionStart := distribution.Range[0]
				distributionEnd := distribution.Range[1]
				if distributionValue >= distributionStart && distributionValue < distributionEnd {
					e.log.Verbose("Bucketing hit allocation and distribution, returning variant %v", distribution.Variant)
					return distribution.Variant
				}
			}
		}
	}
	return segment.Variant
}

func mergeMetadata(metadata []map[string]interface{}) map[string]interface{} {
	mergedMetadata := make(map[string]interface{})
	for _, m := range metadata {
		for k, v := range m {
			mergedMetadata[k] = v
		}
	}
	if len(mergedMetadata) == 0 {
		return nil
	} else {
		return mergedMetadata
	}
}

func matchNull(op string, filterValues []string) bool {
	containsNone := containsNone(filterValues)
	switch op {
	case OpIs, OpContains, OpLessThan, OpLessThanEquals, OpGreaterThan,
		OpGreaterThanEquals, OpVersionLessThan, OpVersionLessThanEquals,
		OpVersionGreaterThan, OpVersionGreaterThanEquals, OpSetIs, OpSetContains,
		OpSetContainsAny:
		return containsNone
	case OpIsNot, OpDoesNotContain, OpSetDoesNotContain, OpSetDoesNotContainAny:
		return !containsNone
	case OpRegexMatch:
		return false
	case OpRegexDoesNotMatch, OpSetIsNot:
		return true
	default:
		return false
	}
}

func matchSet(propValues []string, op string, filterValues []string) bool {
	switch op {
	case OpSetIs:
		return matchesSetIs(propValues, filterValues)
	case OpSetIsNot:
		return !matchesSetIs(propValues, filterValues)
	case OpSetContains:
		return matchesSetContainsAll(propValues, filterValues)
	case OpSetDoesNotContain:
		return !matchesSetContainsAll(propValues, filterValues)
	case OpSetContainsAny:
		return matchesSetContainsAny(propValues, filterValues)
	case OpSetDoesNotContainAny:
		return !matchesSetContainsAny(propValues, filterValues)
	default:
		return false
	}
}

func matchString(propValue string, op string, filterValues []string) bool {
	switch op {
	case OpIs:
		return matchesIs(propValue, filterValues)
	case OpIsNot:
		return !matchesIs(propValue, filterValues)
	case OpContains:
		return matchesContains(propValue, filterValues)
	case OpDoesNotContain:
		return !matchesContains(propValue, filterValues)
	case OpLessThan, OpLessThanEquals, OpGreaterThan, OpGreaterThanEquals:
		return compare(propValue, op, filterValues)
	case OpVersionLessThan, OpVersionLessThanEquals, OpVersionGreaterThan,
		OpVersionGreaterThanEquals:
		return compareVersion(propValue, op, filterValues)
	case OpRegexMatch:
		return matchesRegex(propValue, filterValues)
	case OpRegexDoesNotMatch:
		return !matchesRegex(propValue, filterValues)
	default:
		return false
	}
}

func matchesIs(propValue string, filterValues []string) bool {
	if containsBooleans(filterValues) {
		propValueLower := strings.ToLower(propValue)
		if propValueLower == "true" || propValueLower == "false" {
			for _, filterValue := range filterValues {
				filterValueLower := strings.ToLower(filterValue)
				if propValueLower == filterValueLower {
					return true
				}
			}
		}
	}
	for _, filterValue := range filterValues {
		if filterValue == propValue {
			return true
		}
	}
	return false
}

func matchesContains(propValue string, filterValues []string) bool {
	for _, filterValue := range filterValues {
		propValueLower := strings.ToLower(propValue)
		filterValueLower := strings.ToLower(filterValue)
		if strings.Contains(propValueLower, filterValueLower) {
			return true
		}
	}
	return false
}

func matchesSetIs(propValues, filterValues []string) bool {
	if propValues == nil && filterValues == nil {
		return true
	} else if propValues == nil || filterValues == nil {
		return false
	}
	m1 := make(map[string]bool)
	m2 := make(map[string]bool)
	maxLen := len(filterValues)
	if len(propValues) > len(filterValues) {
		maxLen = len(propValues)
	}
	for i := 0; i < maxLen; i++ {
		if i < len(propValues) {
			m1[propValues[i]] = true
		}
		if i < len(filterValues) {
			m2[filterValues[i]] = true
		}
	}
	if len(m1) != len(m2) {
		return false
	}
	for k := range m1 {
		if _, ok := m2[k]; !ok {
			return false
		}
	}
	return true
}

func matchesSetContainsAll(propValues []string, filterValues []string) bool {
	for _, filterValue := range filterValues {
		if !matchesIs(filterValue, propValues) {
			return false
		}
	}
	return true
}

func matchesSetContainsAny(propValues []string, filterValues []string) bool {
	for _, filterValue := range filterValues {
		if matchesIs(filterValue, propValues) {
			return true
		}
	}
	return false
}

func compare(propValue string, op string, filterValues []string) bool {
	// Attempt to parse the propValue as a number
	propValueNumber, err := strconv.ParseFloat(propValue, 64)
	if err == nil {
		// Attempt to parse filterValues as numbers
		var filterValueNumbers []float64
		for _, filterValue := range filterValues {
			filterValueNumber, err := strconv.ParseFloat(filterValue, 64)
			if err != nil {
				continue
			}
			filterValueNumbers = append(filterValueNumbers, filterValueNumber)
		}
		if filterValueNumbers != nil {
			// Prop value and at least one filter value can be compared as numbers
			return compareNumber(propValueNumber, op, filterValueNumbers)
		}
	}

	// Compare strings
	return compareString(propValue, op, filterValues)
}

func compareVersion(propValue string, op string, filterValues []string) bool {
	// Attempt to parse the propValue as a version
	propValueVersion := parseVersion(propValue)
	if propValueVersion == nil {
		// Fall back on string comparison
		return compareString(propValue, op, filterValues)
	}
	// Attempt to parse filterValues as versions
	var filterValueVersions []version
	for _, filterValue := range filterValues {
		filterValueVersion := parseVersion(filterValue)
		if filterValueVersion == nil {
			continue
		}
		filterValueVersions = append(filterValueVersions, *filterValueVersion)
	}
	if filterValueVersions == nil {
		// Fall back on string comparison
		return compareString(propValue, op, filterValues)
	}
	// Prop value and at least one filter value can be compared as versions
	for _, filterValueVersion := range filterValueVersions {
		compareResult := versionCompare(*propValueVersion, filterValueVersion)
		var result bool
		switch op {
		case OpVersionLessThan:
			result = compareResult < 0
		case OpVersionLessThanEquals:
			result = compareResult <= 0
		case OpVersionGreaterThan:
			result = compareResult > 0
		case OpVersionGreaterThanEquals:
			result = compareResult >= 0
		default:
			result = false
		}
		if result {
			return true
		}
	}
	return false
}

func compareString(propValue string, op string, filterValues []string) bool {
	for _, filterValue := range filterValues {
		var result bool
		switch op {
		case OpLessThan:
			result = propValue < filterValue
		case OpLessThanEquals:
			result = propValue <= filterValue
		case OpGreaterThan:
			result = propValue > filterValue
		case OpGreaterThanEquals:
			result = propValue >= filterValue
		default:
			result = false
		}
		if result {
			return true
		}
	}
	return false
}

func compareNumber(propValue float64, op string, filterValues []float64) bool {
	for _, filterValue := range filterValues {
		var result bool
		switch op {
		case OpLessThan:
			result = propValue < filterValue
		case OpLessThanEquals:
			result = propValue <= filterValue
		case OpGreaterThan:
			result = propValue > filterValue
		case OpGreaterThanEquals:
			result = propValue >= filterValue
		default:
			result = false
		}
		if result {
			return true
		}
	}
	return false
}

func matchesRegex(propValue string, filterValues []string) bool {
	for _, filterValue := range filterValues {
		match, _ := regexp.MatchString(filterValue, propValue)
		if match {
			return true
		}
	}
	return false
}

func containsNone(filterValues []string) bool {
	for _, filterValue := range filterValues {
		if filterValue == "(none)" {
			return true
		}
	}
	return false
}

func containsBooleans(filterValues []string) bool {
	for _, filterValue := range filterValues {
		filterValueLower := strings.ToLower(filterValue)
		if filterValueLower == "true" || filterValueLower == "false" {
			return true
		}
	}
	return false
}

func coerceString(value interface{}) *string {
	if value == nil {
		return nil
	}
	kind := reflect.TypeOf(value).Kind()
	if kind == reflect.Map || kind == reflect.Slice || kind == reflect.Array {
		b, err := json.Marshal(value)
		if err == nil {
			s := string(b)
			return &s
		}
	}
	s := fmt.Sprintf("%v", value)
	return &s
}

func coerceStringList(value interface{}) ([]string, error) {
	// Convert a list to a list of strings
	_, ok := value.([]string)
	if ok {
		return value.([]string), nil
	}
	// Fall back to reflection for slices with unexpected value types
	kind := reflect.TypeOf(value).Kind()
	if kind == reflect.Slice || kind == reflect.Array {
		var stringList []string
		sliceValue := reflect.ValueOf(value)
		for i := 0; i < sliceValue.Len(); i++ {
			e := sliceValue.Index(i).Interface()
			stringValue := coerceString(e)
			if stringValue != nil {
				stringList = append(stringList, *stringValue)
			}
		}
		return stringList, nil
	} else {
		// Parse a string as json array and convert to list of strings, or
		// return null if the string could not be parsed as a json array.
		stringValue := fmt.Sprintf("%v", value)
		var stringList []string
		err := json.Unmarshal([]byte(stringValue), &stringList)
		if err != nil {
			return nil, err
		}
		return stringList, nil
	}
}

func isSetOperator(op string) bool {
	switch op {
	case OpSetIs, OpSetIsNot, OpSetContains, OpSetDoesNotContain,
		OpSetContainsAny, OpSetDoesNotContainAny:
		return true
	default:
		return false
	}
}
