package local

type HTTPErrorResponseException struct {
	StatusCode int
	Message    string
}

func (e *HTTPErrorResponseException) Error() string {
	return e.Message
}

type CohortTooLargeException struct {
	Message string
}

func (e *CohortTooLargeException) Error() string {
	return e.Message
}

type CohortNotModifiedException struct {
	Message string
}

func (e *CohortNotModifiedException) Error() string {
	return e.Message
}
