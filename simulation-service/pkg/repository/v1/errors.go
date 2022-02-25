package v1

import "errors"

var (
	// ErrNoData is a comman error when we are not able to find any data in db we should give this
	ErrNoData = errors.New("no data found")
)
