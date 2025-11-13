package local

// EvaluateOptions contains options for evaluating variants for a user.
type EvaluateOptions struct {
	// FlagKeys are the flags to evaluate with the user. If nil or empty, all flags are evaluated.
	FlagKeys []string
	// TracksExposure indicates whether to track exposure event for the evaluation.
	TracksExposure *bool
}
