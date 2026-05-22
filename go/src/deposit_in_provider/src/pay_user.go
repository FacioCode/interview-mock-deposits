package deposit_in_provider

import (
	"github.com/sirupsen/logrus"
	deposit "interview_mock_deposits_go/deposit/src"
)

func PayUser(dep deposit.Deposit, requestId string) error {
	logrus.WithField("depositId", dep.DepositId).Info("paying user via partner_api")
	// TODO: implement real payment logic
	return nil
}
