package local

import (
	"reflect"
	"testing"
)

func TestEmpty(t *testing.T) {
	// No flag keys
	{
		inputFlags := flagsArray()
		inputFlagKeys := make([]string, 0)
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray()
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag keys
	{
		inputFlags := flagsArray()
		inputFlagKeys := []string{"1"}
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray()
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
}

func TestSingleFlagNoDependencies(t *testing.T) {
	// No flag keys
	{
		inputFlags := flagsArray(flag{Key: "1", Dependencies: []string{}})
		inputFlagKeys := make([]string, 0)
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray(flag{Key: "1", Dependencies: []string{}})
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag keys
	{
		inputFlags := flagsArray(flag{Key: "1", Dependencies: []string{}})
		inputFlagKeys := []string{"1"}
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray(flag{Key: "1", Dependencies: []string{}})
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag no match
	{
		inputFlags := flagsArray(flag{Key: "1", Dependencies: []string{}})
		inputFlagKeys := []string{"999"}
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray()
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
}

func TestSingleFlagWithDependencies(t *testing.T) {
	// No flag keys
	{
		inputFlags := flagsArray(flag{Key: "1", Dependencies: []string{"2"}})
		inputFlagKeys := make([]string, 0)
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray(flag{Key: "1", Dependencies: []string{"2"}})
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag keys
	{
		inputFlags := flagsArray(flag{Key: "1", Dependencies: []string{"2"}})
		inputFlagKeys := []string{"1"}
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray(flag{Key: "1", Dependencies: []string{"2"}})
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag no match
	{
		inputFlags := flagsArray(flag{Key: "1", Dependencies: []string{"2"}})
		inputFlagKeys := []string{"999"}
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray()
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
}
func TestMultipleFlagsNoDependencies(t *testing.T) {
	// No flag keys
	{
		inputFlags := flagsArray(
			flag{Key: "1", Dependencies: []string{}},
			flag{Key: "2", Dependencies: []string{}})
		inputFlagKeys := make([]string, 0)
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray(
			flag{Key: "1", Dependencies: []string{}},
			flag{Key: "2", Dependencies: []string{}})
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag keys
	{
		inputFlags := flagsArray(
			flag{Key: "1", Dependencies: []string{}},
			flag{Key: "2", Dependencies: []string{}})
		inputFlagKeys := []string{"1", "2"}
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray(
			flag{Key: "1", Dependencies: []string{}},
			flag{Key: "2", Dependencies: []string{}})
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag no match
	{
		inputFlags := flagsArray(
			flag{Key: "1", Dependencies: []string{}},
			flag{Key: "2", Dependencies: []string{}})
		inputFlagKeys := []string{"99", "999"}
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray()
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
}
func TestMultipleFlagWithDependencies(t *testing.T) {
	// No flag keys
	{
		inputFlags := flagsArray(
			flag{Key: "1", Dependencies: []string{"2"}},
			flag{Key: "2", Dependencies: []string{"3"}},
			flag{Key: "3", Dependencies: []string{}})
		inputFlagKeys := make([]string, 0)
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray(
			flag{Key: "3", Dependencies: []string{}},
			flag{Key: "2", Dependencies: []string{"3"}},
			flag{Key: "1", Dependencies: []string{"2"}})
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag keys
	{
		inputFlags := flagsArray(
			flag{Key: "1", Dependencies: []string{"2"}},
			flag{Key: "2", Dependencies: []string{"3"}},
			flag{Key: "3", Dependencies: []string{}})
		inputFlagKeys := []string{"1", "2"}
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray(
			flag{Key: "3", Dependencies: []string{}},
			flag{Key: "2", Dependencies: []string{"3"}},
			flag{Key: "1", Dependencies: []string{"2"}})
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag no match
	{
		inputFlags := flagsArray(
			flag{Key: "1", Dependencies: []string{"2"}},
			flag{Key: "2", Dependencies: []string{"3"}},
			flag{Key: "3", Dependencies: []string{}})
		inputFlagKeys := []string{"999"}
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray()
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
}
func TestSingleFlagCycle(t *testing.T) {
	// No flag keys
	{
		inputFlags := flagsArray(flag{Key: "1", Dependencies: []string{"1"}})
		inputFlagKeys := make([]string, 0)
		_, err := topologicalSortArray(inputFlags, inputFlagKeys)
		if err == nil {
			t.Fatalf("expected cycle error")
		}
	}
	// With flag keys
	{
		inputFlags := flagsArray(flag{Key: "1", Dependencies: []string{"1"}})
		inputFlagKeys := []string{"1"}
		_, err := topologicalSortArray(inputFlags, inputFlagKeys)
		if err == nil {
			t.Fatalf("expected cycle error")
		}

	}
	// With flag no match
	{
		inputFlags := flagsArray(flag{Key: "1", Dependencies: []string{"1"}})
		inputFlagKeys := []string{"999"}
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray()
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
}
func TestTwoFlagCycle(t *testing.T) {
	// No flag keys
	{
		inputFlags := flagsArray(
			flag{Key: "1", Dependencies: []string{"2"}},
			flag{Key: "2", Dependencies: []string{"1"}})
		inputFlagKeys := make([]string, 0)
		_, err := topologicalSortArray(inputFlags, inputFlagKeys)
		if err == nil {
			t.Fatalf("expected cycle error")
		}
	}
	// With flag keys
	{
		inputFlags := flagsArray(
			flag{Key: "1", Dependencies: []string{"2"}},
			flag{Key: "2", Dependencies: []string{"1"}})
		inputFlagKeys := []string{"2"}
		_, err := topologicalSortArray(inputFlags, inputFlagKeys)
		if err == nil {
			t.Fatalf("expected cycle error")
		}
	}
	// With flag no match
	{
		inputFlags := flagsArray(
			flag{Key: "1", Dependencies: []string{"2"}},
			flag{Key: "2", Dependencies: []string{"1"}})
		inputFlagKeys := []string{"999"}
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray()
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
}
func TestMultipleFlagsComplexCycle(t *testing.T) {
	inputFlags := flagsArray(
		flag{Key: "3", Dependencies: []string{"1", "2"}},
		flag{Key: "1", Dependencies: []string{}},
		flag{Key: "4", Dependencies: []string{"21", "3"}},
		flag{Key: "2", Dependencies: []string{}},
		flag{Key: "5", Dependencies: []string{"3"}},
		flag{Key: "6", Dependencies: []string{}},
		flag{Key: "7", Dependencies: []string{}},
		flag{Key: "8", Dependencies: []string{"9"}},
		flag{Key: "9", Dependencies: []string{}},
		flag{Key: "20", Dependencies: []string{"4"}},
		flag{Key: "21", Dependencies: []string{"20"}},
	)
	inputFlagKeys := make([]string, 0)
	_, err := topologicalSortArray(inputFlags, inputFlagKeys)
	if err == nil {
		t.Fatalf("expected cycle error")
	}
}
func TestComplexNoCycleStartingWithLeaf(t *testing.T) {
	inputFlags := flagsArray(
		flag{Key: "1", Dependencies: []string{"6", "3"}},
		flag{Key: "2", Dependencies: []string{"8", "5", "3", "1"}},
		flag{Key: "3", Dependencies: []string{"6", "5"}},
		flag{Key: "4", Dependencies: []string{"8", "7"}},
		flag{Key: "5", Dependencies: []string{"10", "7"}},
		flag{Key: "7", Dependencies: []string{"8"}},
		flag{Key: "6", Dependencies: []string{"7", "4"}},
		flag{Key: "8", Dependencies: []string{}},
		flag{Key: "9", Dependencies: []string{"10", "7", "5"}},
		flag{Key: "10", Dependencies: []string{"7"}},
		flag{Key: "20", Dependencies: []string{}},
		flag{Key: "21", Dependencies: []string{"20"}},
		flag{Key: "30", Dependencies: []string{}},
	)
	inputFlagKeys := make([]string, 0)
	actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
	expected := flagsArray(
		flag{Key: "8", Dependencies: []string{}},
		flag{Key: "7", Dependencies: []string{"8"}},
		flag{Key: "4", Dependencies: []string{"8", "7"}},
		flag{Key: "6", Dependencies: []string{"7", "4"}},
		flag{Key: "10", Dependencies: []string{"7"}},
		flag{Key: "5", Dependencies: []string{"10", "7"}},
		flag{Key: "3", Dependencies: []string{"6", "5"}},
		flag{Key: "1", Dependencies: []string{"6", "3"}},
		flag{Key: "2", Dependencies: []string{"8", "5", "3", "1"}},
		flag{Key: "9", Dependencies: []string{"10", "7", "5"}},
		flag{Key: "20", Dependencies: []string{}},
		flag{Key: "21", Dependencies: []string{"20"}},
		flag{Key: "30", Dependencies: []string{}},
	)
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}
func TestComplexNoCycleStartingWithMiddle(t *testing.T) {
	inputFlags := flagsArray(
		flag{Key: "6", Dependencies: []string{"7", "4"}},
		flag{Key: "1", Dependencies: []string{"6", "3"}},
		flag{Key: "2", Dependencies: []string{"8", "5", "3", "1"}},
		flag{Key: "3", Dependencies: []string{"6", "5"}},
		flag{Key: "4", Dependencies: []string{"8", "7"}},
		flag{Key: "5", Dependencies: []string{"10", "7"}},
		flag{Key: "7", Dependencies: []string{"8"}},
		flag{Key: "8", Dependencies: []string{}},
		flag{Key: "9", Dependencies: []string{"10", "7", "5"}},
		flag{Key: "10", Dependencies: []string{"7"}},
		flag{Key: "20", Dependencies: []string{}},
		flag{Key: "21", Dependencies: []string{"20"}},
		flag{Key: "30", Dependencies: []string{}},
	)
	inputFlagKeys := make([]string, 0)
	actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
	expected := flagsArray(
		flag{Key: "8", Dependencies: []string{}},
		flag{Key: "7", Dependencies: []string{"8"}},
		flag{Key: "4", Dependencies: []string{"8", "7"}},
		flag{Key: "6", Dependencies: []string{"7", "4"}},
		flag{Key: "10", Dependencies: []string{"7"}},
		flag{Key: "5", Dependencies: []string{"10", "7"}},
		flag{Key: "3", Dependencies: []string{"6", "5"}},
		flag{Key: "1", Dependencies: []string{"6", "3"}},
		flag{Key: "2", Dependencies: []string{"8", "5", "3", "1"}},
		flag{Key: "9", Dependencies: []string{"10", "7", "5"}},
		flag{Key: "20", Dependencies: []string{}},
		flag{Key: "21", Dependencies: []string{"20"}},
		flag{Key: "30", Dependencies: []string{}},
	)
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}
func TestComplexNoCycleStartingWithRoot(t *testing.T) {
	inputFlags := flagsArray(
		flag{Key: "8", Dependencies: []string{}},
		flag{Key: "1", Dependencies: []string{"6", "3"}},
		flag{Key: "2", Dependencies: []string{"8", "5", "3", "1"}},
		flag{Key: "3", Dependencies: []string{"6", "5"}},
		flag{Key: "4", Dependencies: []string{"8", "7"}},
		flag{Key: "5", Dependencies: []string{"10", "7"}},
		flag{Key: "7", Dependencies: []string{"8"}},
		flag{Key: "6", Dependencies: []string{"7", "4"}},
		flag{Key: "9", Dependencies: []string{"10", "7", "5"}},
		flag{Key: "10", Dependencies: []string{"7"}},
		flag{Key: "20", Dependencies: []string{}},
		flag{Key: "21", Dependencies: []string{"20"}},
		flag{Key: "30", Dependencies: []string{}},
	)
	inputFlagKeys := make([]string, 0)
	actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
	expected := flagsArray(
		flag{Key: "8", Dependencies: []string{}},
		flag{Key: "7", Dependencies: []string{"8"}},
		flag{Key: "4", Dependencies: []string{"8", "7"}},
		flag{Key: "6", Dependencies: []string{"7", "4"}},
		flag{Key: "10", Dependencies: []string{"7"}},
		flag{Key: "5", Dependencies: []string{"10", "7"}},
		flag{Key: "3", Dependencies: []string{"6", "5"}},
		flag{Key: "1", Dependencies: []string{"6", "3"}},
		flag{Key: "2", Dependencies: []string{"8", "5", "3", "1"}},
		flag{Key: "9", Dependencies: []string{"10", "7", "5"}},
		flag{Key: "20", Dependencies: []string{}},
		flag{Key: "21", Dependencies: []string{"20"}},
		flag{Key: "30", Dependencies: []string{}},
	)
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}

// Utilities

type flag struct {
	Key          string
	Dependencies []string
}

func flagsArray(flags ...flag) []interface{} {
	result := make([]interface{}, 0)
	for _, f := range flags {
		result = append(result, flagInterface(f))
	}
	return result
}

func flagInterface(flag flag) interface{} {
	dependencyMap := make(map[string]interface{})
	for _, dependency := range flag.Dependencies {
		dependencyMap[dependency] = true
	}
	return map[string]interface{}{
		"flagKey": flag.Key,
		"parentDependencies": map[string]interface{}{
			"flags": dependencyMap,
		},
	}
}

// Used for testing to ensure the correct ordering of iteration.
func topologicalSortArray(flags []interface{}, flagKeys []string) ([]interface{}, error) {
	// Extract keys and create flags map
	keys := make([]string, 0)
	available := make(map[string]interface{})
	for _, flagAny := range flags {
		switch flag := flagAny.(type) {
		case map[string]interface{}:
			switch flagKey := flag["flagKey"].(type) {
			case string:
				keys = append(keys, flagKey)
				available[flagKey] = flag
			}
		}
	}
	// Get the starting keys
	startingKeys := make([]string, 0)
	if len(flagKeys) > 0 {
		startingKeys = flagKeys
	} else {
		startingKeys = keys
	}
	return topologicalSort(available, startingKeys)
}
