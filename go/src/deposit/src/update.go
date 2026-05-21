package deposit

import (
	"errors"
	"interview_mock_deposits_go/deposit/src/domain"
	"interview_mock_deposits_go/utils"
)

type InconsistentStatusChangeError struct {
	OldStatus domain.DepositStatus
	NewStatus domain.DepositStatus
}

func (e InconsistentStatusChangeError) Error() string {
	return "inconsistent status change"
}

func ValidateStatusChange(oldStatus, newStatus domain.DepositStatus) bool {
	// TODO: replace with real state-machine
	return oldStatus != newStatus
}

func ConfirmDeposit(depositId, receiptURL string, authorizationCode *string) error {
	// TODO: replace with real implementation
	return nil
}

func UpdateDepositStatus(depositId string, newStatus domain.DepositStatus) error {
	// TODO: replace with real implementation
	return nil
}

func UpdateDepositStatusWithReason(depositId string, newStatus domain.DepositStatus, reason, statusDescription string) error {
	// TODO: replace with real implementation
	return nil
}

func UpdateDepositAsFailed(depositId, reason, statusDescription, provider string) error {
	// TODO: replace with real implementation
	return nil
}

func ScheduleToRetry(depositId string, retryAt utils.DateTime, reason string) error {
	if depositId == "" {
		return errors.New("invalid depositId")
	}
	// TODO: replace with real implementation
	return nil
}
