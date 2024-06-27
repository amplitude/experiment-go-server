package experiment

import "sync"

const VERSION = "1.5.0"

type User struct {
	UserId             string                         `json:"user_id,omitempty"`
	DeviceId           string                         `json:"device_id,omitempty"`
	Country            string                         `json:"country,omitempty"`
	Region             string                         `json:"region,omitempty"`
	Dma                string                         `json:"dma,omitempty"`
	City               string                         `json:"city,omitempty"`
	Language           string                         `json:"language,omitempty"`
	Platform           string                         `json:"platform,omitempty"`
	Version            string                         `json:"version,omitempty"`
	Os                 string                         `json:"os,omitempty"`
	DeviceManufacturer string                         `json:"device_manufacturer,omitempty"`
	DeviceBrand        string                         `json:"device_brand,omitempty"`
	DeviceModel        string                         `json:"device_model,omitempty"`
	Carrier            string                         `json:"carrier,omitempty"`
	Library            string                         `json:"library,omitempty"`
	UserProperties     map[string]interface{}         `json:"user_properties,omitempty"`
	Groups             map[string][]string            `json:"groups,omitempty"`
	CohortIDs          map[string]struct{}            `json:"cohort_ids,omitempty"`
	GroupCohortIDs     map[string]map[string]struct{} `json:"group_cohort_ids,omitempty"`
	lock               sync.Mutex
}

func (u *User) AddGroupCohortIDs(groupType, groupName string, cohortIDs map[string]struct{}) {
	u.lock.Lock()
	defer u.lock.Unlock()

	if u.GroupCohortIDs == nil {
		u.GroupCohortIDs = make(map[string]map[string]struct{})
	}

	groupNames := u.GroupCohortIDs[groupType]
	if groupNames == nil {
		groupNames = make(map[string]struct{})
		u.GroupCohortIDs[groupType] = groupNames
	}

	for id := range cohortIDs {
		groupNames[id] = struct{}{}
	}
}

type Variant struct {
	Value    string                 `json:"value,omitempty"`
	Payload  interface{}            `json:"payload,omitempty"`
	Key      string                 `json:"key,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
