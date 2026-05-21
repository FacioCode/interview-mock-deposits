package domain

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type TransactionType int

const (
	SalaryAdvance TransactionType = iota
)

var typeToString = map[TransactionType]string{
	SalaryAdvance: "SalaryAdvance",
}

func (t TransactionType) String() string {
	return typeToString[t]
}

var transactionToID = map[string]TransactionType{
	"SalaryAdvance": SalaryAdvance,
}

func StringToType(status string) (TransactionType, error) {
	if val, ok := transactionToID[status]; ok {
		return val, nil
	}
	return 0, errors.New("Invalid TransactionType")
}

func (t TransactionType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(typeToString[t])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (t *TransactionType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*t = transactionToID[j]
	return nil
}

func (t TransactionType) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	value := t.String()
	av.S = &value
	return nil
}

func (t *TransactionType) UnmarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	*t = transactionToID[*av.S]
	return nil
}
