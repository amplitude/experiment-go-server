package local

type evaluationVariant struct {
	Key     string      `json:"key,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

type flagResult struct {
	Variant          evaluationVariant `json:"variant,omitempty"`
	Description      string            `json:"description,omitempty"`
	IsDefaultVariant bool              `json:"isDefaultVariant,omitempty"`
}

type evaluationResult = map[string]flagResult

type interopResult struct {
	Result *evaluationResult `json:"result,omitempty"`
	Error  *string           `json:"error,omitempty"`
}
