package local

import (
	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"sort"
	"strings"
	"time"
)

type exposure struct {
	user      *experiment.User
	results   map[string]experiment.Variant
	timestamp int64
}

func newExposure(user *experiment.User, results map[string]experiment.Variant) *exposure {
	exposure := &exposure{
		user:      user,
		results:   results,
		timestamp: time.Now().UnixMilli(),
	}

	return exposure
}

func (e *exposure) Canonicalize() string {
	var sb strings.Builder

	if e.user != nil {
		sb.WriteString(e.user.UserId)
		sb.WriteString(" ")
		sb.WriteString(e.user.DeviceId)
		sb.WriteString(" ")
	}

	keys := make([]string, 0, len(e.results))
	for key := range e.results {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := e.results[key].Key
		sb.WriteString(key)
		sb.WriteString(" ")
		sb.WriteString(value)
		sb.WriteString(" ")
	}

	return sb.String()
}

