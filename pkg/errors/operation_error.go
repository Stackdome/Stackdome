package errors

import stderrors "errors"

type RetryableError interface {
	error
	IsRetryable() bool
}

func IsRetryable(err error) bool {
	var re RetryableError
	if stderrors.As(err, &re) {
		return re.IsRetryable()
	}
	return false
}

type OperationError struct {
	Reason    string
	Message   string
	Retryable bool
	Err       error
}

func (e *OperationError) Error() string     { return e.Message }
func (e *OperationError) Unwrap() error     { return e.Err }
func (e *OperationError) IsRetryable() bool { return e.Retryable }

func Permanent(reason, message string) *OperationError {
	return &OperationError{Reason: reason, Message: message}
}

func Transient(reason, message string) *OperationError {
	return &OperationError{Reason: reason, Message: message, Retryable: true}
}
