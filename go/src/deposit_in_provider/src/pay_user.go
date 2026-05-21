package deposit_in_provider

import (
	"context"

	deposit "interview_mock_deposits_go/deposit/src"
	"interview_mock_deposits_go/partner_api"

	"github.com/sirupsen/logrus"
)

func PayUser(dep deposit.Deposit, requestId string) error {
	logrus.WithField("depositId", dep.DepositId).Info("paying user via partner_api")

	_, err := partner_api.Pay(context.Background(), partner_api.TransferRequest{
		ID:     dep.DepositId,
		Amount: dep.Amount,
	})
	if err != nil {
		// TODO: replace with real failure handling (retry / mark failed)
		logrus.WithField("error", err.Error()).Warn("partner_api.Pay returned error")
		return nil
	}
	return nil
}
