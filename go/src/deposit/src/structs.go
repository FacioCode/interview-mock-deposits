package deposit

import (
	"errors"
	"interview_mock_deposits_go/deposit/src/domain"
	"interview_mock_deposits_go/internal/testUtils"
	"interview_mock_deposits_go/utils"
)

type Transaction struct {
	UserId        string                 `json:"userId"`
	Type          domain.TransactionType `json:"type"`
	TransactionId string                 `json:"transactionId"`
	Amount        float64                `json:"amount"`
	DelayHours    *int                   `json:"delayHours,omitempty"`
}

type Deposit struct {
	DepositId string               `json:"depositId"`
	Status    domain.DepositStatus `json:"status"`
	Transaction
	UserData    *UserData       `json:"userData,omitempty"`
	DepositTime *utils.DateTime `json:"depositTime,omitempty"`
	CreatedAt   string          `json:"createdAt,omitempty"`
}

type UserData struct {
	Name        string      `json:"name"`
	Document    string      `json:"document"`
	BankAccount BankAccount `json:"bankAccount"`
}

func (t UserData) Validate() error {
	if t.Document == "" {
		return errors.New("invalid document")
	}
	if t.Name == "" {
		return errors.New("invalid name")
	}
	return t.BankAccount.Validate()
}

type BankAccount struct {
	Bank    string `json:"bank"`
	Branch  string `json:"branch"`
	Account string `json:"account"`
}

func (t BankAccount) Validate() error {
	if t.Bank == "" {
		return errors.New("invalid bank")
	}
	if t.Branch == "" {
		return errors.New("invalid bank branch")
	}
	if t.Account == "" {
		return errors.New("invalid bank account")
	}
	return nil
}

func (t Transaction) GenerateIdempotencyKey() string {
	return testUtils.GenerateRandomId()
}

func (t Transaction) GenerateDepositId() string {
	return testUtils.GenerateRandomId()
}
