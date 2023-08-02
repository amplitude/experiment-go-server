package local

import (
	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"sort"
	"strings"
	"time"
)

type assignment struct {
	user      *experiment.User
	results   *evaluationResult
	timestamp int
}

func newAssignment(user *experiment.User, results *evaluationResult) *assignment {
	assignment := &assignment{
		user:      user,
		results:   results,
		timestamp: int(time.Now().UnixNano() / int64(time.Millisecond)),
	}

	return assignment
}

func (a *assignment) Canonicalize() string {
	var sb strings.Builder

	if a.user != nil {
		sb.WriteString(a.user.UserId)
		sb.WriteString(" ")
		sb.WriteString(a.user.DeviceId)
		sb.WriteString(" ")
	}

	keys := make([]string, 0, len(*a.results))
	for key := range *a.results {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := (*a.results)[key].Variant.Key
		sb.WriteString(key)
		sb.WriteString(" ")
		sb.WriteString(value)
		sb.WriteString(" ")
	}

	return sb.String()
}
