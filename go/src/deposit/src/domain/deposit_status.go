package domain

import (
	"bytes"
	"encoding/json"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type DepositStatus int

const (
	New DepositStatus = iota
	DepositSent
	Canceled
	Failed
	Waiting
	Reversed
)

var statusToString = map[DepositStatus]string{
	New:         "NEW",
	DepositSent: "DEPOSIT_SENT",
	Canceled:    "CANCELED",
	Failed:      "FAILED",
	Waiting:     "WAITING",
	Reversed:    "REVERSED",
}

func (d DepositStatus) String() string {
	return statusToString[d]
}

var statusToID = map[string]DepositStatus{
	"NEW":          New,
	"DEPOSIT_SENT": DepositSent,
	"CANCELED":     Canceled,
	"FAILED":       Failed,
	"WAITING":      Waiting,
	"REVERSED":     Reversed,
}

func StringToStatus(status string) DepositStatus {
	return statusToID[status]
}

func (d DepositStatus) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(statusToString[d])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (d *DepositStatus) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*d = statusToID[j]
	return nil
}

func (d DepositStatus) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	value := d.String()
	av.S = &value
	return nil
}

func (d *DepositStatus) UnmarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	*d = statusToID[*av.S]
	return nil
}
