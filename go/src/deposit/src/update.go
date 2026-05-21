package deposit

import (
	"interview_mock_deposits_go/deposit/src/domain"
)

type InconsistentStatusChangeError struct {
	OldStatus domain.DepositStatus
	NewStatus domain.DepositStatus
}

func (e InconsistentStatusChangeError) Error() string {
	return "inconsistent status change"
}

func ValidateStatusChange(oldStatus, newStatus domain.DepositStatus) bool {
	// TODO: replace with real state-machine validation if needed, currently just checks that status is changing
	return oldStatus != newStatus
}

func ConfirmDeposit(depositId, receiptURL string, authorizationCode *string) error {
	// TODO: replace with real implementation if needed
	return nil
}

func UpdateDepositStatus(depositId string, newStatus domain.DepositStatus) error {
	// TODO: replace with real implementation if needed
	return nil
}

func UpdateDepositStatusWithReason(depositId string, newStatus domain.DepositStatus, reason, statusDescription string) error {
	// TODO: replace with real implementation if needed, otherwise just call UpdateDepositStatus
	return nil
}

func UpdateDepositAsFailed(depositId, reason, statusDescription, provider string) error {
	// TODO: replace with real implementation if needed
	return nil
}
