package deposit

import (
	"errors"
	"time"
)

const (
	LockTTLMinutes = 5
	LockPrefix     = "lock-"
)

var ErrLockAlreadyHeld = errors.New("lock already held by another process")
var ErrInvalidDepositId = errors.New("depositId is required")

var GetNowFunc = time.Now

func AcquireIdempotencyLock(depositId string) error {
	if depositId == "" {
		return ErrInvalidDepositId
	}
	// TODO: replace with real DynamoDB conditional write
	return nil
}
