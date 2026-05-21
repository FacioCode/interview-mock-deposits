package events

const (
	Source               = "interview.mock.deposits"
	DepositRequestedType = "deposit-requested"
)

type DepositRequestedEvent struct {
	Id string `json:"id"`
}
