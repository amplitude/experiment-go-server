package remote

type FetchOptions struct {
	// FlagKeys limits the fetched variants to the given flag keys.
	// When nil, variants for all flags are requested.
	FlagKeys []string
	// TracksAssignment indicates whether to track assignment event for the fetch.
	// Default is true, which means the assignment event will be tracked.
	TracksAssignment bool
	// TracksExposure indicates whether to track exposure event for the fetch.
	// Default is false, which means the exposure event will not be tracked.
	TracksExposure bool
}

var DefaultFetchOptions = &FetchOptions{
	TracksAssignment: true,
	TracksExposure:   false,
}
