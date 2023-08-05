package local

import (
	"fmt"
	"github.com/amplitude/analytics-go/amplitude"
)

const DayMillis = 24 * 60 * 60 * 1000
const FlagTypeMutualExclusionGroup = "mutual-exclusion-group"
const FlagTypeHoldoutGroup = "mutual-holdout-group"

type AssignmentService struct {
	Amplitude *amplitude.Client
	Filter    *assignmentFilter
}

func (s *AssignmentService) Track(assignment *assignment) {
	if s.Filter.shouldTrack(assignment) {
		(*s.Amplitude).Track(toEvent(assignment))
	}
}

func toEvent(assignment *assignment) amplitude.Event {

	event := amplitude.Event{
		EventType:       "[Experiment] assignment",
		UserID:          assignment.user.UserId,
		DeviceID:        assignment.user.DeviceId,
		EventProperties: make(map[string]interface{}),
		UserProperties:  make(map[amplitude.IdentityOp]map[string]interface{}),
	}

	// Loop to set event_properties
	for resultsKey, result := range *assignment.results {
		event.EventProperties[fmt.Sprintf("%s.variant", resultsKey)] = result.Variant.Key
	}

	set := make(map[string]interface{})
	unset := make(map[string]interface{})

	// Loop to set user_properties
	for resultsKey, result := range *assignment.results {
		if result.Type == FlagTypeMutualExclusionGroup {
			continue
		} else if result.IsDefaultVariant {
			unset[fmt.Sprintf("[Experiment] %s", resultsKey)] = "-"
		} else {
			set[fmt.Sprintf("[Experiment] %s", resultsKey)] = result.Variant.Key
		}
	}

	event.UserProperties["$set"] = set
	event.UserProperties["$unset"] = unset

	event.InsertID = fmt.Sprintf("%s %s %d %d", event.UserID, event.DeviceID, hashCode(assignment.Canonicalize()), assignment.timestamp/DayMillis)
	return event
}
