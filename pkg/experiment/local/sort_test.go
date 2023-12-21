package local

import (
	"github.com/amplitude/experiment-go-server/internal/evaluation"
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
		inputFlags := flagsArray(evaluation.Flag{Key: "1", Dependencies: []string{}})
		inputFlagKeys := make([]string, 0)
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray(evaluation.Flag{Key: "1", Dependencies: []string{}})
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag keys
	{
		inputFlags := flagsArray(evaluation.Flag{Key: "1", Dependencies: []string{}})
		inputFlagKeys := []string{"1"}
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray(evaluation.Flag{Key: "1", Dependencies: []string{}})
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag no match
	{
		inputFlags := flagsArray(evaluation.Flag{Key: "1", Dependencies: []string{}})
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
		inputFlags := flagsArray(evaluation.Flag{Key: "1", Dependencies: []string{"2"}})
		inputFlagKeys := make([]string, 0)
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray(evaluation.Flag{Key: "1", Dependencies: []string{"2"}})
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag keys
	{
		inputFlags := flagsArray(evaluation.Flag{Key: "1", Dependencies: []string{"2"}})
		inputFlagKeys := []string{"1"}
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray(evaluation.Flag{Key: "1", Dependencies: []string{"2"}})
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag no match
	{
		inputFlags := flagsArray(evaluation.Flag{Key: "1", Dependencies: []string{"2"}})
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
			evaluation.Flag{Key: "1", Dependencies: []string{}},
			evaluation.Flag{Key: "2", Dependencies: []string{}})
		inputFlagKeys := make([]string, 0)
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray(
			evaluation.Flag{Key: "1", Dependencies: []string{}},
			evaluation.Flag{Key: "2", Dependencies: []string{}})
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag keys
	{
		inputFlags := flagsArray(
			evaluation.Flag{Key: "1", Dependencies: []string{}},
			evaluation.Flag{Key: "2", Dependencies: []string{}})
		inputFlagKeys := []string{"1", "2"}
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray(
			evaluation.Flag{Key: "1", Dependencies: []string{}},
			evaluation.Flag{Key: "2", Dependencies: []string{}})
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag no match
	{
		inputFlags := flagsArray(
			evaluation.Flag{Key: "1", Dependencies: []string{}},
			evaluation.Flag{Key: "2", Dependencies: []string{}})
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
			evaluation.Flag{Key: "1", Dependencies: []string{"2"}},
			evaluation.Flag{Key: "2", Dependencies: []string{"3"}},
			evaluation.Flag{Key: "3", Dependencies: []string{}})
		inputFlagKeys := make([]string, 0)
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray(
			evaluation.Flag{Key: "3", Dependencies: []string{}},
			evaluation.Flag{Key: "2", Dependencies: []string{"3"}},
			evaluation.Flag{Key: "1", Dependencies: []string{"2"}})
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag keys
	{
		inputFlags := flagsArray(
			evaluation.Flag{Key: "1", Dependencies: []string{"2"}},
			evaluation.Flag{Key: "2", Dependencies: []string{"3"}},
			evaluation.Flag{Key: "3", Dependencies: []string{}})
		inputFlagKeys := []string{"1", "2"}
		actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
		expected := flagsArray(
			evaluation.Flag{Key: "3", Dependencies: []string{}},
			evaluation.Flag{Key: "2", Dependencies: []string{"3"}},
			evaluation.Flag{Key: "1", Dependencies: []string{"2"}})
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected %v, actual %v", expected, actual)
		}
	}
	// With flag no match
	{
		inputFlags := flagsArray(
			evaluation.Flag{Key: "1", Dependencies: []string{"2"}},
			evaluation.Flag{Key: "2", Dependencies: []string{"3"}},
			evaluation.Flag{Key: "3", Dependencies: []string{}})
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
		inputFlags := flagsArray(evaluation.Flag{Key: "1", Dependencies: []string{"1"}})
		inputFlagKeys := make([]string, 0)
		_, err := topologicalSortArray(inputFlags, inputFlagKeys)
		if err == nil {
			t.Fatalf("expected cycle error")
		}
	}
	// With flag keys
	{
		inputFlags := flagsArray(evaluation.Flag{Key: "1", Dependencies: []string{"1"}})
		inputFlagKeys := []string{"1"}
		_, err := topologicalSortArray(inputFlags, inputFlagKeys)
		if err == nil {
			t.Fatalf("expected cycle error")
		}

	}
	// With flag no match
	{
		inputFlags := flagsArray(evaluation.Flag{Key: "1", Dependencies: []string{"1"}})
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
			evaluation.Flag{Key: "1", Dependencies: []string{"2"}},
			evaluation.Flag{Key: "2", Dependencies: []string{"1"}})
		inputFlagKeys := make([]string, 0)
		_, err := topologicalSortArray(inputFlags, inputFlagKeys)
		if err == nil {
			t.Fatalf("expected cycle error")
		}
	}
	// With flag keys
	{
		inputFlags := flagsArray(
			evaluation.Flag{Key: "1", Dependencies: []string{"2"}},
			evaluation.Flag{Key: "2", Dependencies: []string{"1"}})
		inputFlagKeys := []string{"2"}
		_, err := topologicalSortArray(inputFlags, inputFlagKeys)
		if err == nil {
			t.Fatalf("expected cycle error")
		}
	}
	// With flag no match
	{
		inputFlags := flagsArray(
			evaluation.Flag{Key: "1", Dependencies: []string{"2"}},
			evaluation.Flag{Key: "2", Dependencies: []string{"1"}})
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
		evaluation.Flag{Key: "3", Dependencies: []string{"1", "2"}},
		evaluation.Flag{Key: "1", Dependencies: []string{}},
		evaluation.Flag{Key: "4", Dependencies: []string{"21", "3"}},
		evaluation.Flag{Key: "2", Dependencies: []string{}},
		evaluation.Flag{Key: "5", Dependencies: []string{"3"}},
		evaluation.Flag{Key: "6", Dependencies: []string{}},
		evaluation.Flag{Key: "7", Dependencies: []string{}},
		evaluation.Flag{Key: "8", Dependencies: []string{"9"}},
		evaluation.Flag{Key: "9", Dependencies: []string{}},
		evaluation.Flag{Key: "20", Dependencies: []string{"4"}},
		evaluation.Flag{Key: "21", Dependencies: []string{"20"}},
	)
	inputFlagKeys := make([]string, 0)
	_, err := topologicalSortArray(inputFlags, inputFlagKeys)
	if err == nil {
		t.Fatalf("expected cycle error")
	}
}
func TestComplexNoCycleStartingWithLeaf(t *testing.T) {
	inputFlags := flagsArray(
		evaluation.Flag{Key: "1", Dependencies: []string{"6", "3"}},
		evaluation.Flag{Key: "2", Dependencies: []string{"8", "5", "3", "1"}},
		evaluation.Flag{Key: "3", Dependencies: []string{"6", "5"}},
		evaluation.Flag{Key: "4", Dependencies: []string{"8", "7"}},
		evaluation.Flag{Key: "5", Dependencies: []string{"10", "7"}},
		evaluation.Flag{Key: "7", Dependencies: []string{"8"}},
		evaluation.Flag{Key: "6", Dependencies: []string{"7", "4"}},
		evaluation.Flag{Key: "8", Dependencies: []string{}},
		evaluation.Flag{Key: "9", Dependencies: []string{"10", "7", "5"}},
		evaluation.Flag{Key: "10", Dependencies: []string{"7"}},
		evaluation.Flag{Key: "20", Dependencies: []string{}},
		evaluation.Flag{Key: "21", Dependencies: []string{"20"}},
		evaluation.Flag{Key: "30", Dependencies: []string{}},
	)
	inputFlagKeys := make([]string, 0)
	actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
	expected := flagsArray(
		evaluation.Flag{Key: "8", Dependencies: []string{}},
		evaluation.Flag{Key: "7", Dependencies: []string{"8"}},
		evaluation.Flag{Key: "4", Dependencies: []string{"8", "7"}},
		evaluation.Flag{Key: "6", Dependencies: []string{"7", "4"}},
		evaluation.Flag{Key: "10", Dependencies: []string{"7"}},
		evaluation.Flag{Key: "5", Dependencies: []string{"10", "7"}},
		evaluation.Flag{Key: "3", Dependencies: []string{"6", "5"}},
		evaluation.Flag{Key: "1", Dependencies: []string{"6", "3"}},
		evaluation.Flag{Key: "2", Dependencies: []string{"8", "5", "3", "1"}},
		evaluation.Flag{Key: "9", Dependencies: []string{"10", "7", "5"}},
		evaluation.Flag{Key: "20", Dependencies: []string{}},
		evaluation.Flag{Key: "21", Dependencies: []string{"20"}},
		evaluation.Flag{Key: "30", Dependencies: []string{}},
	)
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}
func TestComplexNoCycleStartingWithMiddle(t *testing.T) {
	inputFlags := flagsArray(
		evaluation.Flag{Key: "6", Dependencies: []string{"7", "4"}},
		evaluation.Flag{Key: "1", Dependencies: []string{"6", "3"}},
		evaluation.Flag{Key: "2", Dependencies: []string{"8", "5", "3", "1"}},
		evaluation.Flag{Key: "3", Dependencies: []string{"6", "5"}},
		evaluation.Flag{Key: "4", Dependencies: []string{"8", "7"}},
		evaluation.Flag{Key: "5", Dependencies: []string{"10", "7"}},
		evaluation.Flag{Key: "7", Dependencies: []string{"8"}},
		evaluation.Flag{Key: "8", Dependencies: []string{}},
		evaluation.Flag{Key: "9", Dependencies: []string{"10", "7", "5"}},
		evaluation.Flag{Key: "10", Dependencies: []string{"7"}},
		evaluation.Flag{Key: "20", Dependencies: []string{}},
		evaluation.Flag{Key: "21", Dependencies: []string{"20"}},
		evaluation.Flag{Key: "30", Dependencies: []string{}},
	)
	inputFlagKeys := make([]string, 0)
	actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
	expected := flagsArray(
		evaluation.Flag{Key: "8", Dependencies: []string{}},
		evaluation.Flag{Key: "7", Dependencies: []string{"8"}},
		evaluation.Flag{Key: "4", Dependencies: []string{"8", "7"}},
		evaluation.Flag{Key: "6", Dependencies: []string{"7", "4"}},
		evaluation.Flag{Key: "10", Dependencies: []string{"7"}},
		evaluation.Flag{Key: "5", Dependencies: []string{"10", "7"}},
		evaluation.Flag{Key: "3", Dependencies: []string{"6", "5"}},
		evaluation.Flag{Key: "1", Dependencies: []string{"6", "3"}},
		evaluation.Flag{Key: "2", Dependencies: []string{"8", "5", "3", "1"}},
		evaluation.Flag{Key: "9", Dependencies: []string{"10", "7", "5"}},
		evaluation.Flag{Key: "20", Dependencies: []string{}},
		evaluation.Flag{Key: "21", Dependencies: []string{"20"}},
		evaluation.Flag{Key: "30", Dependencies: []string{}},
	)
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}
func TestComplexNoCycleStartingWithRoot(t *testing.T) {
	inputFlags := flagsArray(
		evaluation.Flag{Key: "8", Dependencies: []string{}},
		evaluation.Flag{Key: "1", Dependencies: []string{"6", "3"}},
		evaluation.Flag{Key: "2", Dependencies: []string{"8", "5", "3", "1"}},
		evaluation.Flag{Key: "3", Dependencies: []string{"6", "5"}},
		evaluation.Flag{Key: "4", Dependencies: []string{"8", "7"}},
		evaluation.Flag{Key: "5", Dependencies: []string{"10", "7"}},
		evaluation.Flag{Key: "7", Dependencies: []string{"8"}},
		evaluation.Flag{Key: "6", Dependencies: []string{"7", "4"}},
		evaluation.Flag{Key: "9", Dependencies: []string{"10", "7", "5"}},
		evaluation.Flag{Key: "10", Dependencies: []string{"7"}},
		evaluation.Flag{Key: "20", Dependencies: []string{}},
		evaluation.Flag{Key: "21", Dependencies: []string{"20"}},
		evaluation.Flag{Key: "30", Dependencies: []string{}},
	)
	inputFlagKeys := make([]string, 0)
	actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
	expected := flagsArray(
		evaluation.Flag{Key: "8", Dependencies: []string{}},
		evaluation.Flag{Key: "7", Dependencies: []string{"8"}},
		evaluation.Flag{Key: "4", Dependencies: []string{"8", "7"}},
		evaluation.Flag{Key: "6", Dependencies: []string{"7", "4"}},
		evaluation.Flag{Key: "10", Dependencies: []string{"7"}},
		evaluation.Flag{Key: "5", Dependencies: []string{"10", "7"}},
		evaluation.Flag{Key: "3", Dependencies: []string{"6", "5"}},
		evaluation.Flag{Key: "1", Dependencies: []string{"6", "3"}},
		evaluation.Flag{Key: "2", Dependencies: []string{"8", "5", "3", "1"}},
		evaluation.Flag{Key: "9", Dependencies: []string{"10", "7", "5"}},
		evaluation.Flag{Key: "20", Dependencies: []string{}},
		evaluation.Flag{Key: "21", Dependencies: []string{"20"}},
		evaluation.Flag{Key: "30", Dependencies: []string{}},
	)
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}

func TestComplexNoCycleWithFlagKeys(t *testing.T) {
	inputFlags := flagsArray(
		evaluation.Flag{Key: "1", Dependencies: []string{"6", "3"}},
		evaluation.Flag{Key: "2", Dependencies: []string{"8", "5", "3", "1"}},
		evaluation.Flag{Key: "3", Dependencies: []string{"6", "5"}},
		evaluation.Flag{Key: "4", Dependencies: []string{"8", "7"}},
		evaluation.Flag{Key: "5", Dependencies: []string{"10", "7"}},
		evaluation.Flag{Key: "7", Dependencies: []string{"8"}},
		evaluation.Flag{Key: "6", Dependencies: []string{"7", "4"}},
		evaluation.Flag{Key: "8", Dependencies: []string{}},
		evaluation.Flag{Key: "9", Dependencies: []string{"10", "7", "5"}},
		evaluation.Flag{Key: "10", Dependencies: []string{"7"}},
		evaluation.Flag{Key: "20", Dependencies: []string{}},
		evaluation.Flag{Key: "21", Dependencies: []string{"20"}},
		evaluation.Flag{Key: "30", Dependencies: []string{}},
	)
	inputFlagKeys := []string{"1"}
	actual, _ := topologicalSortArray(inputFlags, inputFlagKeys)
	expected := flagsArray(
		evaluation.Flag{Key: "8", Dependencies: []string{}},
		evaluation.Flag{Key: "7", Dependencies: []string{"8"}},
		evaluation.Flag{Key: "4", Dependencies: []string{"8", "7"}},
		evaluation.Flag{Key: "6", Dependencies: []string{"7", "4"}},
		evaluation.Flag{Key: "10", Dependencies: []string{"7"}},
		evaluation.Flag{Key: "5", Dependencies: []string{"10", "7"}},
		evaluation.Flag{Key: "3", Dependencies: []string{"6", "5"}},
		evaluation.Flag{Key: "1", Dependencies: []string{"6", "3"}},
	)
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}

// Utilities

func flagsArray(flags ...evaluation.Flag) []*evaluation.Flag {
	result := make([]*evaluation.Flag, 0)
	for i := 0; i < len(flags); i++ {
		f := flags[i]
		result = append(result, &f)
	}
	return result
}

// Used for testing to ensure the correct ordering of iteration.
func topologicalSortArray(flags []*evaluation.Flag, flagKeys []string) ([]*evaluation.Flag, error) {
	// Extract keys and create flags map
	keys := make([]string, 0)
	available := make(map[string]*evaluation.Flag)
	for _, f := range flags {
		keys = append(keys, f.Key)
		available[f.Key] = f
	}
	// Get the starting keys
	var startingKeys []string
	if len(flagKeys) > 0 {
		startingKeys = flagKeys
	} else {
		startingKeys = keys
	}
	return topologicalSort(available, startingKeys)
}
