package local

import (
	"fmt"

	"github.com/amplitude/analytics-go/amplitude"
)

// amplitudeTracker is an interface for tracking amplitude events
// This allows us to mock the amplitude client in tests
type amplitudeTracker interface {
	Track(event amplitude.Event)
}

type exposureService struct {
	amplitude amplitudeTracker
	filter    *exposureFilter
}

func (s *exposureService) Track(exposure *exposure) {
	if s.filter.shouldTrack(exposure) {
		events := toExposureEvents(exposure, s.filter.ttlMillis)
		for _, event := range events {
			s.amplitude.Track(event)
		}
	}
}

func toExposureEvents(exposure *exposure, ttlMillis int64) []amplitude.Event {
	var events []amplitude.Event
	canonicalized := exposure.Canonicalize()

	for flagKey, variant := range exposure.results {
		trackExposure, ok := variant.Metadata["trackExposure"].(bool)
		if !ok {
			trackExposure = true
		}
		if !trackExposure {
			continue
		}

		// Skip default variant exposures
		isDefault, ok := variant.Metadata["default"].(bool)
		if !ok {
			isDefault = false
		}
		if isDefault {
			continue
		}

		// Determine user properties to set and unset.
		set := make(map[string]interface{})
		unset := make(map[string]interface{})
		flagType, _ := variant.Metadata["flagType"].(string)
		if flagType != flagTypeMutualExclusionGroup {
			if variant.Key != "" {
				set[fmt.Sprintf("[Experiment] %s", flagKey)] = variant.Key
			} else if variant.Value != "" {
				set[fmt.Sprintf("[Experiment] %s", flagKey)] = variant.Value
			}
		}

		// Build event properties.
		eventProperties := make(map[string]interface{})
		eventProperties["[Experiment] Flag Key"] = flagKey
		if variant.Key != "" {
			eventProperties["[Experiment] Variant"] = variant.Key
		} else if variant.Value != "" {
			eventProperties["[Experiment] Variant"] = variant.Value
		}
		if variant.Metadata != nil {
			eventProperties["metadata"] = variant.Metadata
		}

		// Build event.
		event := amplitude.Event{
			EventType:       "[Experiment] Exposure",
			UserID:          exposure.user.UserId,
			DeviceID:        exposure.user.DeviceId,
			EventProperties: eventProperties,
			UserProperties: map[amplitude.IdentityOp]map[string]interface{}{
				"$set":   set,
				"$unset": unset,
			},
			EventOptions: amplitude.EventOptions{
				InsertID: fmt.Sprintf("%s %s %d %d", exposure.user.UserId, exposure.user.DeviceId, hashCode(flagKey+" "+canonicalized), exposure.timestamp/ttlMillis),
			},
		}

		if len(exposure.user.Groups) > 0 {
			event.Groups = exposure.user.Groups
		}

		events = append(events, event)
	}

	return events
}
