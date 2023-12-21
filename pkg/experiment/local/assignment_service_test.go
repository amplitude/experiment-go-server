package local

import (
	"fmt"
	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"reflect"
	"testing"
)

func TestToEvent(t *testing.T) {
	user := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	results := map[string]experiment.Variant{
		"flag-key-1": {
			Key: "on",
			Metadata: map[string]interface{}{
				"segmentName": "Segment",
				"version": 13,
			},
		},
		"flag-key-2": {
			Key:      "control",
			Metadata: map[string]interface{}{
				"default": true,
				"segmentName": "All Other Users",
				"version": 12,
			},
		},
	}

	assignment := newAssignment(user, results)
	event := toEvent(assignment)
	canonicalization := "user device flag-key-1 on flag-key-2 control "
	expectedInsertID := fmt.Sprintf("user device %d %d", hashCode(canonicalization), assignment.timestamp/dayMillis)
	if event.UserID != "user" {
		t.Errorf("UserID was %s, expected %s", event.UserID, "user")
	}
	if event.DeviceID != "device" {
		t.Errorf("DeviceID was %s, expected %s", event.DeviceID, "device")
	}
	if len(event.UserProperties) != 2 {
		t.Errorf("Length of UserProperties was %d, expected %d", len(event.UserProperties), 2)
	}
	if len(event.UserProperties["$set"]) != 1 {
		t.Errorf("Length of UserProperties.$set was %d, expected %d", len(event.UserProperties["$set"]), 1)
	}
	setUserProperty := event.UserProperties["$set"]["[Experiment] flag-key-1"]
	if setUserProperty != "on" {
		t.Errorf("Unexpected user property value%d", setUserProperty)
	}
	if len(event.UserProperties["$unset"]) != 1 {
		t.Errorf("Length of UserProperties.$unset was %d, expected %d", len(event.UserProperties["$unset"]), 1)
	}
	unsetUserProperty := event.UserProperties["$unset"]["[Experiment] flag-key-2"]
	if unsetUserProperty != "-" {
		t.Errorf("Unexpected user property value%d", setUserProperty)
	}
	expectedEventProperties := map[string]interface{}{
		"flag-key-1.variant": "on",
		"flag-key-1.details": "v13 rule:Segment",
		"flag-key-2.variant": "control",
		"flag-key-2.details": "v12 rule:All Other Users",
	}
	if !reflect.DeepEqual(expectedEventProperties, event.EventProperties) {
		t.Errorf("Unexpected event properties %v", event.EventProperties)

	}
	if event.InsertID != expectedInsertID {
		t.Errorf("InsertID was %s, expected %s", event.InsertID, expectedInsertID)
	}

}

func TestToEventNoDetails(t *testing.T) {
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
	event := toEvent(assignment)
	canonicalization := "user device flag-key-1 on flag-key-2 control "
	expectedInsertID := fmt.Sprintf("user device %d %d", hashCode(canonicalization), assignment.timestamp/dayMillis)
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
