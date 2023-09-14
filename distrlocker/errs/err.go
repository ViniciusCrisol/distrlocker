package errs

import "errors"

var (
	ErrLockCannotBeAcquired = errors.New("distrlocker: lock cannot be acquired")

	ErrLockCannotBeReleased = errors.New("distrlocker: lock cannot be released")
)
