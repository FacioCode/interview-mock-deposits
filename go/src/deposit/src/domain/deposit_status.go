package domain

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type DepositStatus int

// Add new states here as needed.
const (
	New DepositStatus = iota
	DepositSent
	Done
	Failed
	Returned
)

var statusToString = map[DepositStatus]string{
	New:         "NEW",
	DepositSent: "DEPOSIT_SENT",
	Done:        "DONE",
	Failed:      "FAILED",
	Returned:    "RETURNED",
}

func (d DepositStatus) String() string {
	return statusToString[d]
}

var statusToID = map[string]DepositStatus{
	"NEW":          New,
	"DEPOSIT_SENT": DepositSent,
	"DONE":         Done,
	"FAILED":       Failed,
	"RETURNED":     Returned,
}

func StringToStatus(status string) (DepositStatus, error) {
	if val, ok := statusToID[status]; ok {
		return val, nil
	}
	return 0, errors.New("invalid DepositStatus")
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
	val, ok := statusToID[j]
	if !ok {
		return errors.New("invalid DepositStatus: " + j)
	}
	*d = val
	return nil
}

func (d DepositStatus) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	value := d.String()
	av.S = &value
	return nil
}

func (d *DepositStatus) UnmarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	if av.S == nil {
		return nil
	}
	val, ok := statusToID[*av.S]
	if !ok {
		return errors.New("invalid DepositStatus: " + *av.S)
	}
	*d = val
	return nil
}
