package create_deposit

import (
	"strconv"

	deposit "interview_mock_deposits_go/deposit/src"
	"interview_mock_deposits_go/deposit/src/domain"
	"interview_mock_deposits_go/internal/events"

	"github.com/sirupsen/logrus"
)

func CreateDeposit(event events.PendingTransactionEvent, requestId string) error {
	logrus.WithFields(logrus.Fields{
		"transactionId": event.TransactionId,
		"customerId":    event.CustomerId,
	}).Info("creating deposit for pending transaction")

	amount, err := strconv.ParseFloat(event.Amount, 64)
	if err != nil {
		return err
	}

	transaction := deposit.Transaction{
		UserId:        event.CustomerId,
		Type:          domain.SalaryAdvance,
		TransactionId: event.TransactionId,
		Amount:        amount,
	}

	userData := deposit.UserData{
		Name:     event.Name,
		Document: event.Document,
		BankAccount: deposit.BankAccount{
			Bank:    event.BankAccount.Bank,
			Branch:  event.BankAccount.Branch,
			Account: event.BankAccount.Account,
		},
	}

	depositId, err := deposit.CreateDepositWithUserData(transaction, userData, requestId)
	if err != nil {
		return err
	}

	// mock event publication (EventBridge PutEvents)
	logrus.WithFields(logrus.Fields{
		"detailType": events.DepositRequestedType,
		"depositId":  depositId,
	}).Info("event would be published")

	return nil
}
