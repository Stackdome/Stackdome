package git

import "errors"

var (
	ErrNotFound    = errors.New("not found")
	ErrAuthFailed  = errors.New("authentication failed")
	ErrRateLimited = errors.New("rate limited")
)
