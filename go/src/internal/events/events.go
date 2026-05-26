package events

const (
	// Deposit lifecycle events (published by create_new_deposit after persisting)
	Source               = "interview.mock.deposits"
	DepositRequestedType = "deposit-requested"

	// Upstream events that trigger a new deposit
	PendingTransactionSource = "interview.mock.transactions"
	PendingTransactionType   = "transaction-pending"
)

type DepositRequestedEvent struct {
	Id            string `json:"id"`
	TransactionId string `json:"transactionId"`
	CustomerId    string `json:"customerId"`
}

type PendingTransactionEvent struct {
	TransactionId string      `json:"transactionId"`
	CustomerId    string      `json:"customerId"`
	Amount        string      `json:"amount"`
	Name          string      `json:"name"`
	Document      string      `json:"document"`
	BankAccount   BankAccount `json:"bankAccount"`
}

type BankAccount struct {
	Bank    string `json:"bank"`
	Branch  string `json:"branch"`
	Account string `json:"account"`
}
