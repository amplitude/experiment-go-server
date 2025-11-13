package local

import (
	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"testing"
	"time"
)

func TestSingleExposure(t *testing.T) {
	user := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	results := map[string]experiment.Variant{
		"flag-key-1": {
			Key: "on",
		},
		"flag-key-2": {
			Key: "control",
		},
	}

	exposure := newExposure(user, results)
	filter := newExposureFilter(100)
	if !filter.shouldTrack(exposure) {
		t.Errorf("exposure should be tracked")
	}
}

func TestDuplicateExposure(t *testing.T) {
	user := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	results := map[string]experiment.Variant{
		"flag-key-1": {
			Key: "on",
		},
		"flag-key-2": {
			Key: "control",
		},
	}

	exposure1 := newExposure(user, results)
	exposure2 := newExposure(user, results)
	filter := newExposureFilter(100)
	if !filter.shouldTrack(exposure1) {
		t.Errorf("Exposure1 should be tracked")
	}
	if filter.shouldTrack(exposure2) {
		t.Errorf("Exposure2 should not be tracked")
	}
}

func TestExposureFilterSameUserDifferentResults(t *testing.T) {
	user := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	results1 := map[string]experiment.Variant{
		"flag-key-1": {
			Key: "on",
		},
		"flag-key-2": {
			Key: "control",
		},
	}

	results2 := map[string]experiment.Variant{
		"flag-key-1": {
			Key: "control",
		},
		"flag-key-2": {
			Key: "on",
		},
	}

	exposure1 := newExposure(user, results1)
	exposure2 := newExposure(user, results2)
	filter := newExposureFilter(100)
	if !filter.shouldTrack(exposure1) {
		t.Errorf("Exposure1 should be tracked")
	}
	if !filter.shouldTrack(exposure2) {
		t.Errorf("Exposure2 should be tracked")
	}
}

func TestExposureFilterSameResultsDifferentUser(t *testing.T) {
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
			Key: "control",
		},
	}

	exposure1 := newExposure(user1, results)
	exposure2 := newExposure(user2, results)
	filter := newExposureFilter(100)
	if !filter.shouldTrack(exposure1) {
		t.Errorf("Exposure1 should be tracked")
	}
	if !filter.shouldTrack(exposure2) {
		t.Errorf("Exposure2 should be tracked")
	}
}

func TestExposureFilterEmptyResult(t *testing.T) {
	user1 := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	user2 := &experiment.User{
		UserId:   "different-user",
		DeviceId: "different-device",
	}

	results := map[string]experiment.Variant{}

	exposure1 := newExposure(user1, results)
	exposure2 := newExposure(user1, results)
	exposure3 := newExposure(user2, results)
	filter := newExposureFilter(100)
	if filter.shouldTrack(exposure1) {
		t.Errorf("Exposure1 should not be tracked")
	}
	if filter.shouldTrack(exposure2) {
		t.Errorf("Exposure2 should not be tracked")
	}
	if filter.shouldTrack(exposure3) {
		t.Errorf("Exposure3 should not be tracked")
	}
}

func TestDuplicateExposuresWithDifferentResultOrder(t *testing.T) {
	user := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	results1 := map[string]experiment.Variant{
		"flag-key-1": {
			Key: "on",
		},
		"flag-key-2": {
			Key: "control",
		},
	}

	results2 := map[string]experiment.Variant{
		"flag-key-2": {
			Key: "control",
		},
		"flag-key-1": {
			Key: "on",
		},
	}

	exposure1 := newExposure(user, results1)
	exposure2 := newExposure(user, results2)
	filter := newExposureFilter(100)
	if !filter.shouldTrack(exposure1) {
		t.Errorf("Exposure1 should be tracked")
	}
	if filter.shouldTrack(exposure2) {
		t.Errorf("Exposure2 should not be tracked")
	}
}

func TestExposureFilterLRUReplacement(t *testing.T) {
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
			Key: "control",
		},
	}

	exposure1 := newExposure(user1, results)
	exposure2 := newExposure(user2, results)
	exposure3 := newExposure(user3, results)
	filter := newExposureFilter(2)
	if !filter.shouldTrack(exposure1) {
		t.Errorf("Exposure1 should be tracked")
	}
	if !filter.shouldTrack(exposure2) {
		t.Errorf("Exposure2 should be tracked")
	}
	if !filter.shouldTrack(exposure3) {
		t.Errorf("Exposure3 should be tracked")
	}
	if !filter.shouldTrack(exposure1) {
		t.Errorf("Exposure1 should be tracked")
	}
}

func TestExposureFilterTTLBasedEviction(t *testing.T) {
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
			Key: "control",
		},
	}

	exposure1 := newExposure(user1, results)
	exposure2 := newExposure(user2, results)
	// Create filter with 1 second TTL
	filter := newExposureFilter(100)
	filter.cache.TTL = 1000
	if !filter.shouldTrack(exposure1) {
		t.Errorf("Exposure1 should be tracked")
	}
	time.Sleep(1050 * time.Millisecond)
	if !filter.shouldTrack(exposure1) {
		t.Errorf("Exposure1 should be tracked")
	}
	if !filter.shouldTrack(exposure2) {
		t.Errorf("Exposure2 should be tracked")
	}
	time.Sleep(950 * time.Millisecond)
	if filter.shouldTrack(exposure2) {
		t.Errorf("Exposure2 should not be tracked")
	}
}

