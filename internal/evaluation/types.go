package evaluation

type Flag struct {
	Key          string                 `json:"key,omitempty"`
	Variants     map[string]*Variant    `json:"variants,omitempty"`
	Segments     []*Segment             `json:"segments,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type Variant struct {
	Key      string                 `json:"key,omitempty"`
	Value    interface{}            `json:"value,omitempty"`
	Payload  interface{}            `json:"payload,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type Segment struct {
	Bucket     *Bucket                `json:"bucket,omitempty"`
	Conditions [][]*Condition         `json:"conditions,omitempty"`
	Variant    string                 `json:"variant,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type Bucket struct {
	Selector    []string      `json:"selector,omitempty"`
	Salt        string        `json:"salt,omitempty"`
	Allocations []*Allocation `json:"allocations,omitempty"`
}

type Condition struct {
	Selector []string `json:"selector,omitempty"`
	Op       string   `json:"op,omitempty"`
	Values   []string `json:"values,omitempty"`
}

type Allocation struct {
	Range         []uint64        `json:"range,omitempty"`
	Distributions []*Distribution `json:"distributions,omitempty"`
}

type Distribution struct {
	Variant string   `json:"variant,omitempty"`
	Range   []uint64 `json:"range,omitempty"`
}

const (
	OpIs                       = "is"
	OpIsNot                    = "is not"
	OpContains                 = "contains"
	OpDoesNotContain           = "does not contain"
	OpLessThan                 = "less"
	OpLessThanEquals           = "less or equal"
	OpGreaterThan              = "greater"
	OpGreaterThanEquals        = "greater or equal"
	OpVersionLessThan          = "version less"
	OpVersionLessThanEquals    = "version less or equal"
	OpVersionGreaterThan       = "version greater"
	OpVersionGreaterThanEquals = "version greater or equal"
	OpSetIs                    = "set is"
	OpSetIsNot                 = "set is not"
	OpSetContains              = "set contains"
	OpSetDoesNotContain        = "set does not contain"
	OpSetContainsAny           = "set contains any"
	OpSetDoesNotContainAny     = "set does not contain any"
	OpRegexMatch               = "regex match"
	OpRegexDoesNotMatch        = "regex does not match"
)
