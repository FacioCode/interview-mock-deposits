package deposit

import (
	"errors"
	"os"
	"strings"
	"time"

	"interview_mock_deposits_go/deposit/src/domain"
	"interview_mock_deposits_go/utils/awsutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var TableName = aws.String(os.Getenv("TABLE_NAME"))

var DdbSvc = awsutil.NewDynamoClient()

var ErrDepositAlreadyExists = errors.New("deposit already exists")

func CreateDeposit(request Transaction, requestId string) error {
	return createDeposit(request, nil, requestId, domain.New)
}

func CreateDepositWithUserData(request Transaction, userData UserData, requestId string) error {
	return createDeposit(request, &userData, requestId, domain.New)
}

func CreateDepositWithStatus(request Transaction, requestId string, status domain.DepositStatus) error {
	return createDeposit(request, nil, requestId, status)
}

func createDeposit(request Transaction, userData *UserData, requestId string, status domain.DepositStatus) error {
	newDeposit := Deposit{
		DepositId:   request.GenerateDepositId(),
		Status:      status,
		Transaction: request,
	}
	if userData != nil && userData.Document != "" {
		newDeposit.UserData = userData
	}
	return SaveDeposit(newDeposit, requestId)
}

func SaveDeposit(data Deposit, requestId string) error {
	item, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		return err
	}
	item["depositId"] = &dynamodb.AttributeValue{S: aws.String(data.DepositId)}
	item["createdAt"] = &dynamodb.AttributeValue{S: aws.String(time.Now().UTC().Format(time.RFC3339))}

	_, err = DdbSvc.PutItem(&dynamodb.PutItemInput{
		Item:                item,
		TableName:           TableName,
		ConditionExpression: aws.String("attribute_not_exists(depositId)"),
	})
	if err != nil && strings.Contains(err.Error(), "ConditionalCheckFailedException") {
		return ErrDepositAlreadyExists
	}
	return err
}
