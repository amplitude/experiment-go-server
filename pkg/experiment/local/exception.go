package local

type httpErrorResponseException struct {
	StatusCode int
	Message    string
}

func (e *httpErrorResponseException) Error() string {
	return e.Message
}

type cohortTooLargeException struct {
	Message string
}

func (e *cohortTooLargeException) Error() string {
	return e.Message
}
