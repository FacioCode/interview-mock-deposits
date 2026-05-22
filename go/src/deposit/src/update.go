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

// TODO: implementar máquina de estados real (ex.: NEW -> DEPOSIT_SENT -> DONE/FAILED/RETURNED;
// proibir transições inválidas como FAILED -> DONE). Hoje aceita qualquer mudança != atual.
func ValidateStatusChange(oldStatus, newStatus domain.DepositStatus) bool {
	return oldStatus != newStatus
}

func UpdateDepositStatus(depositId string, newStatus domain.DepositStatus) error {
	// TODO: replace with real implementation
	return nil
}

func UpdateDepositAsFailed(depositId, reason string) error {
	// TODO: replace with real implementation
	return nil
}
