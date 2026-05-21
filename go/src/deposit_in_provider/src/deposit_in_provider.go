package deposit_in_provider

import (
	deposit "interview_mock_deposits_go/deposit/src"
	"interview_mock_deposits_go/deposit/src/domain"
	"interview_mock_deposits_go/internal/events"

	"github.com/sirupsen/logrus"
)

func CreateDepositInProvider(req events.DepositRequestedEvent, requestId string) error {
	depositId := req.Id
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
