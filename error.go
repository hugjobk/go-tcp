package tcp

import "errors"

var (
	ErrClosed  = errors.New("closed")
	ErrTimeout = errTimeout{}
)

type errTimeout struct{}

func (err errTimeout) Error() string {
	return "timeout"
}

func (err errTimeout) Timeout() bool {
	return true
}
