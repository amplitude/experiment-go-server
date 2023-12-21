package local

import (
	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"testing"
	"time"
)

func TestSingleAssignment(t *testing.T) {
	user := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	results := map[string]experiment.Variant{
		"flag-key-1": {
			Key: "on",
		},
		"flag-key-2": {
			Key:      "control",
			Metadata: map[string]interface{}{"default": true},
		},
	}

	assignment := newAssignment(user, results)
	filter := newAssignmentFilter(100)
	if !filter.shouldTrack(assignment) {
		t.Errorf("assignment should be tracked")
	}
}

func TestDuplicateAssignment(t *testing.T) {
	user := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	results := map[string]experiment.Variant{
		"flag-key-1": {
			Key: "on",
		},
		"flag-key-2": {
			Key:      "control",
			Metadata: map[string]interface{}{"default": true},
		},
	}

	assignment1 := newAssignment(user, results)
	assignment2 := newAssignment(user, results)
	filter := newAssignmentFilter(100)
	if !filter.shouldTrack(assignment1) {
		t.Errorf("Assignment1 should be tracked")
	}
	if filter.shouldTrack(assignment2) {
		t.Errorf("Assignment2 should not be tracked")
	}
}

func TestSameUserDifferentResults(t *testing.T) {
	user := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	results1 := map[string]experiment.Variant{
		"flag-key-1": {
			Key: "on",
		},
		"flag-key-2": {
			Key:      "control",
			Metadata: map[string]interface{}{"default": true},
		},
	}

	results2 := map[string]experiment.Variant{
		"flag-key-1": {
			Key:      "control",
			Metadata: map[string]interface{}{"default": true},
		},
		"flag-key-2": {
			Key: "on",
		},
	}

	assignment1 := newAssignment(user, results1)
	assignment2 := newAssignment(user, results2)
	filter := newAssignmentFilter(100)
	if !filter.shouldTrack(assignment1) {
		t.Errorf("Assignment1 should be tracked")
	}
	if !filter.shouldTrack(assignment2) {
		t.Errorf("Assignment2 should be tracked")
	}
}

func TestSameResultsDifferentUser(t *testing.T) {
	user1 := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	user2 := &experiment.User{
		UserId:   "different-user",
		DeviceId: "different-device",
	}

	results := map[string]experiment.Variant{
		"flag-key-1": {
			Key: "on",
		},
		"flag-key-2": {
			Key:      "control",
			Metadata: map[string]interface{}{"default": true},
		},
	}

	assignment1 := newAssignment(user1, results)
	assignment2 := newAssignment(user2, results)
	filter := newAssignmentFilter(100)
	if !filter.shouldTrack(assignment1) {
		t.Errorf("Assignment1 should be tracked")
	}
	if !filter.shouldTrack(assignment2) {
		t.Errorf("Assignment2 should be tracked")
	}
}

func TestEmptyResult(t *testing.T) {
	user1 := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	user2 := &experiment.User{
		UserId:   "different-user",
		DeviceId: "different-device",
	}

	results := map[string]experiment.Variant{}

	assignment1 := newAssignment(user1, results)
	assignment2 := newAssignment(user1, results)
	assignment3 := newAssignment(user2, results)
	filter := newAssignmentFilter(100)
	if filter.shouldTrack(assignment1) {
		t.Errorf("Assignment1 should not be tracked")
	}
	if filter.shouldTrack(assignment2) {
		t.Errorf("Assignment2 should not be tracked")
	}
	if filter.shouldTrack(assignment3) {
		t.Errorf("Assignment3 should not be tracked")
	}
}

func TestDuplicateAssignmentsWithDifferentResultOrder(t *testing.T) {
	user := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	results1 := map[string]experiment.Variant{
		"flag-key-1": {
			Key: "on",
		},
		"flag-key-2": {
			Key:      "control",
			Metadata: map[string]interface{}{"default": true},
		},
	}

	results2 := map[string]experiment.Variant{
		"flag-key-2": {
			Key:      "control",
			Metadata: map[string]interface{}{"default": true},
		},
		"flag-key-1": {
			Key: "on",
		},
	}

	assignment1 := newAssignment(user, results1)
	assignment2 := newAssignment(user, results2)
	filter := newAssignmentFilter(100)
	if !filter.shouldTrack(assignment1) {
		t.Errorf("Assignment1 should be tracked")
	}
	if filter.shouldTrack(assignment2) {
		t.Errorf("Assignment2 should not be tracked")
	}
}

func TestLRUReplacement(t *testing.T) {
	user1 := &experiment.User{
		UserId:   "user1",
		DeviceId: "device",
	}

	user2 := &experiment.User{
		UserId:   "user2",
		DeviceId: "device",
	}

	user3 := &experiment.User{
		UserId:   "user3",
		DeviceId: "device",
	}

	results := map[string]experiment.Variant{
		"flag-key-1": {
			Key: "on",
		},
		"flag-key-2": {
			Key:      "control",
			Metadata: map[string]interface{}{"default": true},
		},
	}

	assignment1 := newAssignment(user1, results)
	assignment2 := newAssignment(user2, results)
	assignment3 := newAssignment(user3, results)
	filter := newAssignmentFilter(2)
	if !filter.shouldTrack(assignment1) {
		t.Errorf("Assignment1 should be tracked")
	}
	if !filter.shouldTrack(assignment2) {
		t.Errorf("Assignment2 should be tracked")
	}
	if !filter.shouldTrack(assignment3) {
		t.Errorf("Assignment3 should be tracked")
	}
	if !filter.shouldTrack(assignment1) {
		t.Errorf("Assignment1 should be tracked")
	}
}

func TestTTLBasedEviction(t *testing.T) {
	user1 := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	user2 := &experiment.User{
		UserId:   "different-user",
		DeviceId: "different-device",
	}

	results := map[string]experiment.Variant{
		"flag-key-1": {
			Key: "on",
		},
		"flag-key-2": {
			Key:      "control",
			Metadata: map[string]interface{}{"default": true},
		},
	}

	assignment1 := newAssignment(user1, results)
	assignment2 := newAssignment(user2, results)
	filter := newAssignmentFilter(100)
	filter.cache.TTL = 1000
	if !filter.shouldTrack(assignment1) {
		t.Errorf("Assignment1 should be tracked")
	}
	time.Sleep(1050 * time.Millisecond)
	if !filter.shouldTrack(assignment1) {
		t.Errorf("Assignment1 should be tracked")
	}
	if !filter.shouldTrack(assignment2) {
		t.Errorf("Assignment2 should be tracked")
	}
	time.Sleep(950 * time.Millisecond)
	if filter.shouldTrack(assignment2) {
		t.Errorf("Assignment2 should not be tracked")
	}
}
