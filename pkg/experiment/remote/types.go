package remote

type interopVariant struct {
	Value   string      `json:"value,omitempty"`
	Key     string      `json:"key,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

type interopVariants = map[string]interopVariant
