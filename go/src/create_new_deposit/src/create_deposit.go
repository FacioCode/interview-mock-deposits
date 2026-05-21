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

	// TODO: replace with real user-data fetch (customer service + bank account lookup)
	userData := deposit.UserData{}

	return deposit.CreateDepositWithUserData(transaction, userData, requestId)
}
