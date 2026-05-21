package webhook

import (
	deposit "interview_mock_deposits_go/deposit/src"
	"interview_mock_deposits_go/deposit/src/domain"

	"github.com/sirupsen/logrus"
)

func HandleWebhook(body Body) error {
	if body.Object != TRANSFER_OBJ {
		return nil
	}

	switch body.Data.Status {
	case DONE:
		return deposit.ConfirmDeposit(body.Data.IntegrationId, body.Data.BankReceiptURL, body.Data.AuthorizationCode)
	case FAILED:
		logrus.WithField("depositId", body.Data.IntegrationId).Warn("deposit failed")
		return deposit.UpdateDepositAsFailed(body.Data.IntegrationId, domain.GenericWebhookErrorReason, "", "")
	case RETURNED:
		logrus.WithField("depositId", body.Data.IntegrationId).Info("deposit returned")
		return deposit.UpdateDepositStatusWithReason(body.Data.IntegrationId, domain.New, domain.GenericWebhookErrorReason, "")
	}
	return nil
}
