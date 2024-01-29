package remote

import "fmt"

type fetchError struct {
	StatusCode int
	Message    string
}

func (e *fetchError) Error() string {
	return fmt.Sprintf("message: %s", e.Message)
}
