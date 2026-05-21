package deposit

import (
	"errors"

	"interview_mock_deposits_go/deposit/src/domain"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func GetDepositByUserId(userId string) ([]Deposit, error) {
	out, err := DdbSvc.Query(&dynamodb.QueryInput{
		TableName: TableName,
		IndexName: aws.String("userIndex"),
		KeyConditions: map[string]*dynamodb.Condition{
			"userId": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{S: aws.String(userId)},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	var deposits []Deposit
	if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &deposits); err != nil {
		return nil, err
	}
	return deposits, nil
}

func GetDepositById(id string) (Deposit, error) {
	if id == "" {
		return Deposit{}, errors.New("Deposit not found")
	}
	// TODO: replace with real implementation
	return Deposit{}, nil
}

func GetDepositByStatus(status domain.DepositStatus) ([]Deposit, error) {
	// TODO: replace with real implementation, if needed
	return []Deposit{}, nil
}
