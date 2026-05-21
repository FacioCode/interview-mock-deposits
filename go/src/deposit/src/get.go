package deposit

import (
	"errors"

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
	out, err := DdbSvc.GetItem(&dynamodb.GetItemInput{
		TableName:      TableName,
		ConsistentRead: aws.Bool(true),
		Key: map[string]*dynamodb.AttributeValue{
			"depositId": {S: aws.String(id)},
		},
	})
	if err != nil {
		return Deposit{}, err
	}
	if out.Item == nil {
		return Deposit{}, errors.New("Deposit not found")
	}
	var d Deposit
	if err := dynamodbattribute.UnmarshalMap(out.Item, &d); err != nil {
		return Deposit{}, err
	}
	return d, nil
}
