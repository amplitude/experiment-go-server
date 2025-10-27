package remote

type FetchOptions struct {
	TracksAssignment bool
	TracksExposure   bool
}

var DefaultFetchOptions = &FetchOptions{
	TracksAssignment: true,
	TracksExposure:   false,
}
