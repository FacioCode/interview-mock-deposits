package deposit

import (
	"interview_mock_deposits_go/deposit/src/domain"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var TableName = aws.String(os.Getenv("TABLE_NAME"))

var DdbSvc *dynamodb.DynamoDB

func init() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	DdbSvc = dynamodb.New(sess)
}

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
	// TODO: replace with real implementation
	_ = request.GenerateDepositId()
	_ = request.GenerateIdempotencyKey()
	return nil
}

func SaveDeposit(data Deposit, requestId string) error {
	// TODO: replace with real implementation
	return nil
}
