package local

type evaluationVariant struct {
	Key     string      `json:"key,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

type flagResult struct {
	Variant evaluationVariant `json:"variant,omitempty"`
}

type evaluationResult = map[string]flagResult
