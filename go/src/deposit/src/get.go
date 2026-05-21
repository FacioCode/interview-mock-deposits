package deposit

import (
	"errors"
	"interview_mock_deposits_go/deposit/src/domain"
)

func GetDepositByStatus(status domain.DepositStatus) ([]Deposit, error) {
	// TODO: replace with real implementation
	return []Deposit{}, nil
}

func GetDepositByUserId(userId string) ([]Deposit, error) {
	// TODO: replace with real implementation
	return []Deposit{}, nil
}

func GetDepositById(id string) (Deposit, error) {
	if id == "" {
		return Deposit{}, errors.New("Deposit not found")
	}
	// TODO: replace with real implementation
	return Deposit{}, nil
}

func GetDepositByIds(depositIds []string) ([]Deposit, error) {
	// TODO: replace with real implementation
	return []Deposit{}, nil
}

func GetDepositByEndToEndId(endToEndId string) (Deposit, error) {
	// TODO: replace with real implementation
	return Deposit{}, nil
}
