package local

import (
	"fmt"
	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"reflect"
	"testing"
)

func TestToExposureEvents(t *testing.T) {
	user := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	results := map[string]experiment.Variant{
		"basic": {
			Key:   "control",
			Value: "control",
			Metadata: map[string]interface{}{
				"segmentName": "All Other Users",
				"flagType":    "experiment",
				"flagVersion": float64(10),
				"default":     false,
			},
		},
		"different_value": {
			Key:   "on",
			Value: "control",
			Metadata: map[string]interface{}{
				"segmentName": "All Other Users",
				"flagType":    "experiment",
				"flagVersion": float64(10),
				"default":     false,
			},
		},
		"default": {
			Key: "off",
			Metadata: map[string]interface{}{
				"segmentName": "All Other Users",
				"flagType":    "experiment",
				"flagVersion": float64(10),
				"default":     true,
			},
		},
		"mutex": {
			Key:   "slot-1",
			Value: "slot-1",
			Metadata: map[string]interface{}{
				"segmentName": "All Other Users",
				"flagType":    "mutual-exclusion-group",
				"flagVersion": float64(10),
				"default":     false,
			},
		},
		"holdout": {
			Key:   "holdout",
			Value: "holdout",
			Metadata: map[string]interface{}{
				"segmentName": "All Other Users",
				"flagType":    "holdout-group",
				"flagVersion": float64(10),
				"default":     false,
			},
		},
		"partial_metadata": {
			Key:   "on",
			Value: "on",
			Metadata: map[string]interface{}{
				"segmentName": "All Other Users",
				"flagType":    "release",
			},
		},
		"empty_metadata": {
			Key:   "on",
			Value: "on",
		},
		"empty_variant": {},
	}

	exposure := newExposure(user, results)
	events := toExposureEvents(exposure, dayMillis)
	// Should exclude default (default=true) only
	// basic, different_value, mutex, holdout, partial_metadata, empty_metadata, empty_variant = 7 events
	if len(events) != 7 {
		t.Errorf("Expected 7 events, got %d", len(events))
	}

	for _, event := range events {
		if event.UserID != "user" {
			t.Errorf("UserID was %s, expected %s", event.UserID, "user")
		}
		if event.DeviceID != "device" {
			t.Errorf("DeviceID was %s, expected %s", event.DeviceID, "device")
		}
		if event.EventType != "[Experiment] Exposure" {
			t.Errorf("EventType was %s, expected %s", event.EventType, "[Experiment] Exposure")
		}

		flagKey := event.EventProperties["[Experiment] Flag Key"].(string)
		variant := results[flagKey]

		// Validate event properties
		if variant.Key != "" {
			if event.EventProperties["[Experiment] Variant"] != variant.Key {
				t.Errorf("Variant was %v, expected %s", event.EventProperties["[Experiment] Variant"], variant.Key)
			}
		}
		if variant.Metadata != nil {
			if !reflect.DeepEqual(event.EventProperties["metadata"], variant.Metadata) {
				t.Errorf("Metadata mismatch")
			}
		}

		// Validate user properties
		flagType, _ := variant.Metadata["flagType"].(string)
		if flagType == "mutual-exclusion-group" {
			if len(event.UserProperties["$set"]) != 0 || len(event.UserProperties["$unset"]) != 0 {
				t.Errorf("Mutual exclusion group should have empty user properties")
			}
		} else {
			if variant.Metadata != nil && variant.Metadata["default"] == true {
				if len(event.UserProperties["$set"]) != 0 {
					t.Errorf("Default variant should not have set properties")
				}
			} else {
				if len(event.UserProperties["$unset"]) != 0 {
					t.Errorf("Non-default variant should not have unset properties")
				}
			}
		}

		// Validate insert id
		canonicalized := exposure.Canonicalize()
		expectedInsertID := fmt.Sprintf("%s %s %d %d", event.UserID, event.DeviceID, hashCode(flagKey+" "+canonicalized), exposure.timestamp/dayMillis)
		if event.EventOptions.InsertID != expectedInsertID {
			t.Errorf("InsertID was %s, expected %s", event.EventOptions.InsertID, expectedInsertID)
		}
	}
}

func TestToExposureEventsSkipsDefault(t *testing.T) {
	user := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	results := map[string]experiment.Variant{
		"flag-key-1": {
			Key: "on",
			Metadata: map[string]interface{}{
				"default": false,
			},
		},
		"flag-key-2": {
			Key: "control",
			Metadata: map[string]interface{}{
				"default": true,
			},
		},
	}

	exposure := newExposure(user, results)
	events := toExposureEvents(exposure, dayMillis)
	// Should exclude default variant
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	if events[0].EventProperties["[Experiment] Flag Key"] != "flag-key-1" {
		t.Errorf("Expected flag-key-1, got %v", events[0].EventProperties["[Experiment] Flag Key"])
	}
}

func TestTrackingCalled(t *testing.T) {
	// Test that tracking is called when filter allows
	// This is a basic test - full integration testing would require mocking amplitude client
	filter := newExposureFilter(100)
	user := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}
	results := map[string]experiment.Variant{
		"flag-key-1": {
			Key: "on",
		},
	}
	exposure := newExposure(user, results)
	// Test that filter allows tracking
	if !filter.shouldTrack(exposure) {
		t.Errorf("Filter should allow tracking")
	}
	// Test that events are generated
	events := toExposureEvents(exposure, dayMillis)
	if len(events) == 0 {
		t.Errorf("Expected events to be generated")
	}
}


