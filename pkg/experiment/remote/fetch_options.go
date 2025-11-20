package remote

type FetchOptions struct {
	// TracksAssignment indicates whether to track assignment event for the fetch.
	// Default is true, which means the assignment event will be tracked.
	TracksAssignment bool
	// TracksExposure indicates whether to track exposure event for the fetch.
	// Default is false, which means the exposure event will not be tracked.
	TracksExposure   bool
}

var DefaultFetchOptions = &FetchOptions{
	TracksAssignment: true,
	TracksExposure:   false,
}
