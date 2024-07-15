package experiment

const VERSION = "1.5.0"

type User struct {
	UserId             string                                    `json:"user_id,omitempty"`
	DeviceId           string                                    `json:"device_id,omitempty"`
	Country            string                                    `json:"country,omitempty"`
	Region             string                                    `json:"region,omitempty"`
	Dma                string                                    `json:"dma,omitempty"`
	City               string                                    `json:"city,omitempty"`
	Language           string                                    `json:"language,omitempty"`
	Platform           string                                    `json:"platform,omitempty"`
	Version            string                                    `json:"version,omitempty"`
	Os                 string                                    `json:"os,omitempty"`
	DeviceManufacturer string                                    `json:"device_manufacturer,omitempty"`
	DeviceBrand        string                                    `json:"device_brand,omitempty"`
	DeviceModel        string                                    `json:"device_model,omitempty"`
	Carrier            string                                    `json:"carrier,omitempty"`
	Library            string                                    `json:"library,omitempty"`
	UserProperties     map[string]interface{}                    `json:"user_properties,omitempty"`
	GroupProperties    map[string]map[string]string              `json:"group_properties,omitempty"`
	Groups             map[string][]string                       `json:"groups,omitempty"`
	CohortIds          map[string]struct{}                       `json:"cohort_ids,omitempty"`
	GroupCohortIds     map[string]map[string]map[string]struct{} `json:"group_cohort_ids,omitempty"`
}

func (u *User) AddGroupCohortIds(groupType, groupName string, cohortIds map[string]struct{}) {
	if u.GroupCohortIds == nil {
		u.GroupCohortIds = make(map[string]map[string]map[string]struct{})
	}

	groupNames := u.GroupCohortIds[groupType]
	if groupNames == nil {
		groupNames = make(map[string]map[string]struct{})
		u.GroupCohortIds[groupType] = groupNames
	}

	groupNames[groupName] = cohortIds
}

type Variant struct {
	Value    string                 `json:"value,omitempty"`
	Payload  interface{}            `json:"payload,omitempty"`
	Key      string                 `json:"key,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
