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

	results := &evaluationResult{
		"flag-key-1": flagResult{
			Variant: evaluationVariant{
				Key: "on",
			},
			IsDefaultVariant: false,
		},
		"flag-key-2": flagResult{
			Variant: evaluationVariant{
				Key: "control",
			},
			IsDefaultVariant: true,
		},
	}

	assignment := NewAssignment(user, results)
	filter := NewAssignmentFilter(100)
	if !filter.shouldTrack(assignment) {
		t.Errorf("Assignment should be tracked")
	}
}

func TestDuplicateAssignment(t *testing.T) {
	user := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	results := &evaluationResult{
		"flag-key-1": flagResult{
			Variant: evaluationVariant{
				Key: "on",
			},
			IsDefaultVariant: false,
		},
		"flag-key-2": flagResult{
			Variant: evaluationVariant{
				Key: "control",
			},
			IsDefaultVariant: true,
		},
	}

	assignment1 := NewAssignment(user, results)
	assignment2 := NewAssignment(user, results)
	filter := NewAssignmentFilter(100)
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

	results1 := &evaluationResult{
		"flag-key-1": flagResult{
			Variant: evaluationVariant{
				Key: "on",
			},
			IsDefaultVariant: false,
		},
		"flag-key-2": flagResult{
			Variant: evaluationVariant{
				Key: "control",
			},
			IsDefaultVariant: true,
		},
	}

	results2 := &evaluationResult{
		"flag-key-1": flagResult{
			Variant: evaluationVariant{
				Key: "control",
			},
			IsDefaultVariant: true,
		},
		"flag-key-2": flagResult{
			Variant: evaluationVariant{
				Key: "on",
			},
			IsDefaultVariant: false,
		},
	}

	assignment1 := NewAssignment(user, results1)
	assignment2 := NewAssignment(user, results2)
	filter := NewAssignmentFilter(100)
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

	results := &evaluationResult{
		"flag-key-1": flagResult{
			Variant: evaluationVariant{
				Key: "on",
			},
			IsDefaultVariant: false,
		},
		"flag-key-2": flagResult{
			Variant: evaluationVariant{
				Key: "control",
			},
			IsDefaultVariant: true,
		},
	}

	assignment1 := NewAssignment(user1, results)
	assignment2 := NewAssignment(user2, results)
	filter := NewAssignmentFilter(100)
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

	results := &evaluationResult{}

	assignment1 := NewAssignment(user1, results)
	assignment2 := NewAssignment(user1, results)
	assignment3 := NewAssignment(user2, results)
	filter := NewAssignmentFilter(100)
	if !filter.shouldTrack(assignment1) {
		t.Errorf("Assignment1 should be tracked")
	}
	if filter.shouldTrack(assignment2) {
		t.Errorf("Assignment2 should not be tracked")
	}
	if !filter.shouldTrack(assignment3) {
		t.Errorf("Assignment3 should be tracked")
	}
}

func TestDuplicateAssignmentsWithDifferentResultOrder(t *testing.T) {
	user := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	results1 := &evaluationResult{
		"flag-key-1": flagResult{
			Variant: evaluationVariant{
				Key: "on",
			},
			IsDefaultVariant: false,
		},
		"flag-key-2": flagResult{
			Variant: evaluationVariant{
				Key: "control",
			},
			IsDefaultVariant: true,
		},
	}

	results2 := &evaluationResult{
		"flag-key-2": flagResult{
			Variant: evaluationVariant{
				Key: "control",
			},
			IsDefaultVariant: true,
		},
		"flag-key-1": flagResult{
			Variant: evaluationVariant{
				Key: "on",
			},
			IsDefaultVariant: false,
		},
	}

	assignment1 := NewAssignment(user, results1)
	assignment2 := NewAssignment(user, results2)
	filter := NewAssignmentFilter(100)
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

	results := &evaluationResult{
		"flag-key-1": flagResult{
			Variant: evaluationVariant{
				Key: "on",
			},
			IsDefaultVariant: false,
		},
		"flag-key-2": flagResult{
			Variant: evaluationVariant{
				Key: "control",
			},
			IsDefaultVariant: true,
		},
	}

	assignment1 := NewAssignment(user1, results)
	assignment2 := NewAssignment(user2, results)
	assignment3 := NewAssignment(user3, results)
	filter := NewAssignmentFilter(2)
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

	results := &evaluationResult{
		"flag-key-1": flagResult{
			Variant: evaluationVariant{
				Key: "on",
			},
			IsDefaultVariant: false,
		},
		"flag-key-2": flagResult{
			Variant: evaluationVariant{
				Key: "control",
			},
			IsDefaultVariant: true,
		},
	}

	assignment1 := NewAssignment(user1, results)
	assignment2 := NewAssignment(user2, results)
	filter := NewAssignmentFilter(100)
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
