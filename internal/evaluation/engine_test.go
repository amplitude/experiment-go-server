package evaluation

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/amplitude/experiment-go-server/logger"
	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	//"os"
	"testing"
)

const deploymentKey = "server-NgJxxvg8OGwwBsWVXqyxQbdiflbhvugy"

var flags []*Flag
var engine = &Engine{logger.New(logger.Error, logger.NewDefault())}

func init() {
	rawFlags, err := getFlagConfigsRaw()
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(rawFlags, &flags)
	if err != nil {
		panic(err)
	}
}

// Basic Tests

func TestOff(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_id":   "user_id",
		"device_id": "device_id",
	})
	result := engine.Evaluate(user, flags)["test-off"]
	if result.Key != "off" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestOn(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_id":   "user_id",
		"device_id": "device_id",
	})
	result := engine.Evaluate(user, flags)["test-on"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

// Opinionated Segment Tests

func TestIndividualInclusionsMatch(t *testing.T) {
	// Match User ID
	user := userContext(map[string]interface{}{
		"user_id": "user_id",
	})
	result := engine.Evaluate(user, flags)["test-individual-inclusions"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
	if result.Metadata["segmentName"] != "individual-inclusions" {
		t.Fatalf("unexpected segment result %v", result.Metadata["segmentName"])
	}
	// Match Device ID
	user = userContext(map[string]interface{}{
		"device_id": "device_id",
	})
	result = engine.Evaluate(user, flags)["test-individual-inclusions"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
	if result.Metadata["segmentName"] != "individual-inclusions" {
		t.Fatalf("unexpected segment result %v", result.Metadata["segmentName"])
	}
	// Doesn't Match User ID
	user = userContext(map[string]interface{}{
		"user_id": "not_user_id",
	})
	result = engine.Evaluate(user, flags)["test-individual-inclusions"]
	if result.Key != "off" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
	// Doesn't Match Device ID
	user = userContext(map[string]interface{}{
		"device_id": "not_device_id",
	})
	result = engine.Evaluate(user, flags)["test-individual-inclusions"]
	if result.Key != "off" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestFlagDependenciesOn(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_id":   "user_id",
		"device_id": "device_id",
	})
	result := engine.Evaluate(user, flags)["test-flag-dependencies-on"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestFlagDependenciesOff(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_id":   "user_id",
		"device_id": "device_id",
	})
	result := engine.Evaluate(user, flags)["test-flag-dependencies-off"]
	if result.Key != "off" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
	if result.Metadata["segmentName"] != "flag-dependencies" {
		t.Fatalf("unexpected segment result %v", result.Metadata["segmentName"])
	}
}

func TestStickyBucketing(t *testing.T) {
	// On
	user := userContext(map[string]interface{}{
		"user_id":   "user_id",
		"device_id": "device_id",
		"user_properties": map[string]interface{}{
			"[Experiment] test-sticky-bucketing": "on",
		},
	})

	result := engine.Evaluate(user, flags)["test-sticky-bucketing"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
	if result.Metadata["segmentName"] != "sticky-bucketing" {
		t.Fatalf("unexpected segment result %v", result.Metadata["segmentName"])
	}
	// Off
	user = userContext(map[string]interface{}{
		"user_id":   "user_id",
		"device_id": "device_id",
		"user_properties": map[string]interface{}{
			"[Experiment] test-sticky-bucketing": "off",
		},
	})
	result = engine.Evaluate(user, flags)["test-sticky-bucketing"]
	if result.Key != "off" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
	if result.Metadata["segmentName"] != "All Other Users" {
		t.Fatalf("unexpected segment result %v", result.Metadata["segmentName"])
	}
	// Non-variant
	user = userContext(map[string]interface{}{
		"user_id":   "user_id",
		"device_id": "device_id",
		"user_properties": map[string]interface{}{
			"[Experiment] test-sticky-bucketing": "not-a-variant",
		},
	})
	result = engine.Evaluate(user, flags)["test-sticky-bucketing"]
	if result.Key != "off" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
	if result.Metadata["segmentName"] != "All Other Users" {
		t.Fatalf("unexpected segment result %v", result.Metadata["segmentName"])
	}
}

// Experiment and Flag Segment Tests

func TestExperiment(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_id":   "user_id",
		"device_id": "device_id",
	})
	result := engine.Evaluate(user, flags)["test-experiment"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
	if result.Metadata["experimentKey"] != "exp-1" {
		t.Fatalf("unexpected experiment key result %v", result.Metadata["experimentKey"])
	}
}

func TestFlag(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_id":   "user_id",
		"device_id": "device_id",
	})
	result := engine.Evaluate(user, flags)["test-flag"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
	if result.Metadata["experimentKey"] != nil {
		t.Fatalf("unexpected experiment key result %v", result.Metadata["experimentKey"])
	}
}

// Conditional Logic Tests

func TestMultipleConditionsAndValues(t *testing.T) {
	// All match, on
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key-1": "value-1",
			"key-2": "value-2",
			"key-3": "value-3",
		},
	})
	result := engine.Evaluate(user, flags)["test-multiple-conditions-and-values"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
	// Some match, off
	user = userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key-1": "value-1",
			"key-2": "value-2",
		},
	})
	result = engine.Evaluate(user, flags)["test-multiple-conditions-and-values"]
	if result.Key != "off" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

// Condition Property Targeting Tests

func TestAmplitudePropertyTargeting(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_id": "user_id",
	})
	result := engine.Evaluate(user, flags)["test-amplitude-property-targeting"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestCohortTargeting(t *testing.T) {
	// User in cohort
	user := userContext(map[string]interface{}{
		"cohort_ids": []string{"u0qtvwla", "12345678"},
	})
	result := engine.Evaluate(user, flags)["test-cohort-targeting"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}

	// User not in cohort
	user = userContext(map[string]interface{}{
		"cohort_ids": []string{"12345678", "87654321"},
	})
	result = engine.Evaluate(user, flags)["test-cohort-targeting"]
	if result.Key != "off" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestGroupNameTargeting(t *testing.T) {
	user := groupContext(map[string]interface{}{
		"org name": map[string]interface{}{
			"group_name": "amplitude",
		},
	})
	result := engine.Evaluate(user, flags)["test-group-name-targeting"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestGroupPropertyTargeting(t *testing.T) {
	user := groupContext(map[string]interface{}{
		"org name": map[string]interface{}{
			"group_name": "amplitude",
			"group_properties": map[string]interface{}{
				"org plan": "enterprise2",
			},
		},
	})
	result := engine.Evaluate(user, flags)["test-group-property-targeting"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

// Bucketing Tests

func TestAmplitudeIdBucketing(t *testing.T) {
	user := userContext(map[string]interface{}{
		"amplitude_id": "1234567890",
	})
	result := engine.Evaluate(user, flags)["test-amplitude-id-bucketing"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestUserIdBucketing(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_id": "user_id",
	})
	result := engine.Evaluate(user, flags)["test-user-id-bucketing"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestDeviceIdBucketing(t *testing.T) {
	user := userContext(map[string]interface{}{
		"device_id": "device_id",
	})
	result := engine.Evaluate(user, flags)["test-device-id-bucketing"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestCustomUserPropertyBucketing(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key": "value",
		},
	})
	result := engine.Evaluate(user, flags)["test-custom-user-property-bucketing"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestGroupNameBucketing(t *testing.T) {
	user := groupContext(map[string]interface{}{
		"org name": map[string]interface{}{
			"group_name": "amplitude",
		},
	})
	result := engine.Evaluate(user, flags)["test-group-name-bucketing"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestGroupPropertyBucketing(t *testing.T) {
	user := groupContext(map[string]interface{}{
		"org name": map[string]interface{}{
			"group_name": "amplitude",
			"group_properties": map[string]interface{}{
				"org plan": "enterprise2",
			},
		},
	})
	result := engine.Evaluate(user, flags)["test-group-property-bucketing"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

// Bucketing Allocation Tests

func Test1PercentAllocation(t *testing.T) {
	on := 0
	for i := 0; i < 10000; i++ {
		user := userContext(map[string]interface{}{
			"device_id": strconv.Itoa(i + 1),
		})
		result := engine.Evaluate(user, flags)["test-1-percent-allocation"]
		if result.Key == "on" {
			on++
		}
	}
	if on != 107 {
		t.Fatalf("unexpected evaluation result %v", on)
	}
}

func Test50PercentAllocation(t *testing.T) {
	on := 0
	for i := 0; i < 10000; i++ {
		user := userContext(map[string]interface{}{
			"device_id": strconv.Itoa(i + 1),
		})
		result := engine.Evaluate(user, flags)["test-50-percent-allocation"]
		if result.Key == "on" {
			on++
		}
	}
	if on != 5009 {
		t.Fatalf("unexpected evaluation result %v", on)
	}
}

func Test99PercentAllocation(t *testing.T) {
	on := 0
	for i := 0; i < 10000; i++ {
		user := userContext(map[string]interface{}{
			"device_id": strconv.Itoa(i + 1),
		})
		result := engine.Evaluate(user, flags)["test-99-percent-allocation"]
		if result.Key == "on" {
			on++
		}
	}
	if on != 9900 {
		t.Fatalf("unexpected evaluation result %v", on)
	}
}

// Bucketing Distribution Tests

func Test1PercentDistribution(t *testing.T) {
	control := 0
	treatment := 0
	for i := 0; i < 10000; i++ {
		user := userContext(map[string]interface{}{
			"device_id": strconv.Itoa(i + 1),
		})
		result := engine.Evaluate(user, flags)["test-1-percent-distribution"]
		switch result.Key {
		case "control":
			control++
		case "treatment":
			treatment++
		default:
			t.Fatalf("unexpected variant %v", result.Key)
		}
	}
	if control != 106 {
		t.Fatalf("unexpected evaluation result %v", control)
	}
	if treatment != 9894 {
		t.Fatalf("unexpected evaluation result %v", treatment)
	}
}

func Test50PercentDistribution(t *testing.T) {
	control := 0
	treatment := 0
	for i := 0; i < 10000; i++ {
		user := userContext(map[string]interface{}{
			"device_id": strconv.Itoa(i + 1),
		})
		result := engine.Evaluate(user, flags)["test-50-percent-distribution"]
		switch result.Key {
		case "control":
			control++
		case "treatment":
			treatment++
		default:
			t.Fatalf("unexpected variant %v", result.Key)
		}
	}
	if control != 4990 {
		t.Fatalf("unexpected evaluation result %v", control)
	}
	if treatment != 5010 {
		t.Fatalf("unexpected evaluation result %v", treatment)
	}
}

func Test99PercentDistribution(t *testing.T) {
	control := 0
	treatment := 0
	for i := 0; i < 10000; i++ {
		user := userContext(map[string]interface{}{
			"device_id": strconv.Itoa(i + 1),
		})
		result := engine.Evaluate(user, flags)["test-99-percent-distribution"]
		switch result.Key {
		case "control":
			control++
		case "treatment":
			treatment++
		default:
			t.Fatalf("unexpected variant %v", result.Key)
		}
	}
	if control != 9909 {
		t.Fatalf("unexpected evaluation result %v", control)
	}
	if treatment != 91 {
		t.Fatalf("unexpected evaluation result %v", treatment)
	}
}

func TestMultipleDistributions(t *testing.T) {
	a := 0
	b := 0
	c := 0
	d := 0
	for i := 0; i < 10000; i++ {
		user := userContext(map[string]interface{}{
			"device_id": strconv.Itoa(i + 1),
		})
		result := engine.Evaluate(user, flags)["test-multiple-distributions"]
		switch result.Key {
		case "a":
			a++
		case "b":
			b++
		case "c":
			c++
		case "d":
			d++
		default:
			t.Fatalf("unexpected variant %v", result.Key)
		}
	}
	if a != 2444 {
		t.Fatalf("unexpected evaluation result %v", a)
	}
	if b != 2634 {
		t.Fatalf("unexpected evaluation result %v", b)
	}
	if c != 2447 {
		t.Fatalf("unexpected evaluation result %v", c)
	}
	if d != 2475 {
		t.Fatalf("unexpected evaluation result %v", d)
	}
}

// Operator Tests

func TestIs(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key": "value",
		},
	})
	result := engine.Evaluate(user, flags)["test-is"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestIsNot(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key": "value",
		},
	})
	result := engine.Evaluate(user, flags)["test-is-not"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestContains(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key": "value",
		},
	})
	result := engine.Evaluate(user, flags)["test-contains"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestDoesNotContain(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key": "value",
		},
	})
	result := engine.Evaluate(user, flags)["test-does-not-contain"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestLess(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key": "-1",
		},
	})
	result := engine.Evaluate(user, flags)["test-less"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestLessOrEqual(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key": "0",
		},
	})
	result := engine.Evaluate(user, flags)["test-less-or-equal"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestGreater(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key": "1",
		},
	})
	result := engine.Evaluate(user, flags)["test-greater"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestGreaterOrEqual(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key": "0",
		},
	})
	result := engine.Evaluate(user, flags)["test-greater-or-equal"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestVersionLess(t *testing.T) {
	user := userContext(map[string]interface{}{
		"version": "1.9.0",
	})
	result := engine.Evaluate(user, flags)["test-version-less"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestVersionLessOrEqual(t *testing.T) {
	user := userContext(map[string]interface{}{
		"version": "1.10.0",
	})
	result := engine.Evaluate(user, flags)["test-version-less-or-equal"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestVersionGreater(t *testing.T) {
	user := userContext(map[string]interface{}{
		"version": "1.10.0",
	})
	result := engine.Evaluate(user, flags)["test-version-greater"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestVersionGreaterOrEqual(t *testing.T) {
	user := userContext(map[string]interface{}{
		"version": "1.9.0",
	})
	result := engine.Evaluate(user, flags)["test-version-greater-or-equal"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestSetIs(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key": []int{1, 2, 3},
		},
	})
	result := engine.Evaluate(user, flags)["test-set-is"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestSetIsNot(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key": []int{1, 2},
		},
	})
	result := engine.Evaluate(user, flags)["test-set-is-not"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestSetContains(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key": []int{1, 2, 3, 4},
		},
	})
	result := engine.Evaluate(user, flags)["test-set-contains"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestSetDoesNotContain(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key": []int{1, 2, 4},
		},
	})
	result := engine.Evaluate(user, flags)["test-set-does-not-contain"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestSetContainsAny(t *testing.T) {
	user := userContext(map[string]interface{}{
		"cohort_ids": []string{"u0qtvwla", "12345678"},
	})
	result := engine.Evaluate(user, flags)["test-set-contains-any"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestSetDoesNotContainAny(t *testing.T) {
	user := userContext(map[string]interface{}{
		"cohort_ids": []string{"12345678", "87654321"},
	})
	result := engine.Evaluate(user, flags)["test-set-does-not-contain-any"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestGlobMatch(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key": "/path/1/2/3/end",
		},
	})
	result := engine.Evaluate(user, flags)["test-glob-match"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

func TestGlobDoesNotMatch(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"key": "/path/1/2/3",
		},
	})
	result := engine.Evaluate(user, flags)["test-glob-does-not-match"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

// Test specific functionality

func TestIsWithBooleans(t *testing.T) {
	user := userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"true":  "TRUE",
			"false": "FALSE",
		},
	})
	result := engine.Evaluate(user, flags)["test-is-with-booleans"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
	user = userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"true":  "True",
			"false": "False",
		},
	})
	result = engine.Evaluate(user, flags)["test-is-with-booleans"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
	user = userContext(map[string]interface{}{
		"user_properties": map[string]interface{}{
			"true":  "true",
			"false": "false",
		},
	})
	result = engine.Evaluate(user, flags)["test-is-with-booleans"]
	if result.Key != "on" {
		t.Fatalf("unexpected evaluation result %v", result.Key)
	}
}

// Util

func userContext(u map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"user": u}
}

func groupContext(g map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"groups": g}
}

func getFlagConfigsRaw() ([]byte, error) {
	client := &http.Client{}
	endpoint, err := url.Parse("https://api.lab.amplitude.com/")
	if err != nil {
		return nil, err
	}
	endpoint.Path = "sdk/v2/flags"
	endpoint.RawQuery = "eval_mode=remote"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", deploymentKey))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Amp-Exp-Library", fmt.Sprintf("experiment-go-server/%v", experiment.VERSION))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
