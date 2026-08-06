package backend

import "fmt"

type HTTPError struct {
	msg  string
	code int
	body string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("status code %d: %s",
		e.code,
		e.msg,
	)
}

func (e *HTTPError) Code() int {
	return e.code
}

func (e *HTTPError) Body() string {
	return e.body
}
