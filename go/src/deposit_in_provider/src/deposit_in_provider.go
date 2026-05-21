package deposit_in_provider

import (
	"fmt"

	deposit "interview_mock_deposits_go/deposit/src"
	"interview_mock_deposits_go/deposit/src/domain"
	"interview_mock_deposits_go/internal/events"

	"github.com/sirupsen/logrus"
)

func CreateDepositInProvider(req events.DepositRequestedEvent, requestId string) error {
	depositId := req.Id

	if err := deposit.AcquireIdempotencyLock(depositId); err != nil {
		if err == deposit.ErrLockAlreadyHeld {
			logrus.WithField("depositId", depositId).Info("duplicate event blocked by idempotency lock")
			return nil
		}
		return fmt.Errorf("failed to acquire idempotency lock: %w", err)
	}

	dep, err := deposit.GetDepositById(depositId)
	if err != nil {
		logrus.WithField("error", err.Error()).Warn("error getting deposit")
		return err
	}
	if dep.Status != domain.New {
		logrus.WithFields(logrus.Fields{"depositId": depositId, "status": dep.Status}).Warn("deposit not new")
		return nil
	}
	if dep.UserData == nil {
		logrus.WithField("depositId", depositId).Error("deposit user data is nil")
		return nil
	}

	return PayUser(dep, requestId)
}
