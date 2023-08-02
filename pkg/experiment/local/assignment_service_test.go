package local

import (
	"fmt"
	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"testing"
)

func TestToEvent(t *testing.T) {
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

	assignment := newAssignment(user, results)
	event := toEvent(assignment)
	canonicalization := "user device flag-key-1 on flag-key-2 control "
	expectedInsertID := fmt.Sprintf("user device %d %d", hashCode(canonicalization), assignment.timestamp/DayMillis)
	if event.UserID != "user" {
		t.Errorf("UserID was %s, expected %s", event.UserID, "user")
	}
	if event.DeviceID != "device" {
		t.Errorf("DeviceID was %s, expected %s", event.DeviceID, "device")
	}
	if len(event.UserProperties) != 2 {
		t.Errorf("Length of UserProperties was %d, expected %d", len(event.UserProperties), 2)
	}
	if len(event.EventProperties) != 2 {
		t.Errorf("Length of EventProperties was %d, expected %d", len(event.EventProperties), 2)
	}
	if event.InsertID != expectedInsertID {
		t.Errorf("InsertID was %s, expected %s", event.InsertID, expectedInsertID)
	}
}
