package parallelize

import (
	"context"
)

// ErrorChannel supports non-blocking send and receive operation to capture error.
// A maximum of one error is kept in the channel and the rest of the errors sent
// are ignored, unless the existing error is received and the channel becomes empty
// again.
type ErrorChannel struct {
	errCh chan error
}

func NewErrorChannel() *ErrorChannel {
	return &ErrorChannel{
		errCh: make(chan error, 1),
	}
}

func (e *ErrorChannel) SendErrorWithCancel(err error, cancel context.CancelFunc) {
	e.SendError(err)
	cancel()
}

func (e *ErrorChannel) SendError(err error) {
	select {
	case e.errCh <- err:
	default:
	}
}

func (e *ErrorChannel) ReceiveError() error {
	select {
	case err := <-e.errCh:
		return err
	default:
		return nil
	}
}
