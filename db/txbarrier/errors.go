package txbarrier

import "errors"

var (
	// ErrDuplicationOrSuspension indicates a duplicated or hanging request occurred.
	ErrDuplicationOrSuspension = errors.New("txbarrier: duplicated or hanging request")
	// ErrEmptyCompensation indicates the empty compensation occurred.
	ErrEmptyCompensation = errors.New("txbarrier: empty compensation")
)
