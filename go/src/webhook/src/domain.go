package webhook

const (
	DONE         = "DONE"
	FAILED       = "FAILED"
	RETURNED     = "RETURNED"
	TRANSFER_OBJ = "Transfer"
)

type Body struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Date    string `json:"date"`
	Data    Data   `json:"data"`
	Version string `json:"version"`
}

type Data struct {
	Id                string  `json:"id"`
	Status            string  `json:"status"`
	StatusDescription *string `json:"status_description"`
	IntegrationId     string  `json:"integration_id"`
	BankReceiptURL    string  `json:"bank_receipt_url"`
	AuthorizationCode *string `json:"authorization_code,omitempty"`
}
