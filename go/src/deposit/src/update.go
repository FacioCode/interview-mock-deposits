package deposit

import (
	"interview_mock_deposits_go/deposit/src/domain"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type InconsistentStatusChangeError struct {
	OldStatus domain.DepositStatus
	NewStatus domain.DepositStatus
}

func (e InconsistentStatusChangeError) Error() string {
	return "inconsistent status change"
}

var validStatusTransitions = map[domain.DepositStatus]map[domain.DepositStatus]bool{
	domain.New: {
		domain.DepositSent: true,
		domain.Failed:      true,
	},
	domain.DepositSent: {
		domain.Done:     true,
		domain.Failed:   true,
		domain.Returned: true,
	},
	domain.Done: {
		domain.Returned: true,
	},
	domain.Failed:   {},
	domain.Returned: {},
}

func ValidateStatusChange(oldStatus, newStatus domain.DepositStatus) bool {
	if oldStatus == newStatus {
		return false
	}
	allowed, ok := validStatusTransitions[oldStatus]
	if !ok {
		return false
	}
	return allowed[newStatus]
}

func UpdateDepositStatus(depositId string, newStatus domain.DepositStatus) error {
	current, err := GetDepositById(depositId)
	if err != nil {
		return err
	}
	if !ValidateStatusChange(current.Status, newStatus) {
		return InconsistentStatusChangeError{OldStatus: current.Status, NewStatus: newStatus}
	}
	_, err = DdbSvc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: TableName,
		Key: map[string]*dynamodb.AttributeValue{
			"depositId": {S: aws.String(depositId)},
		},
		UpdateExpression: aws.String("SET #status = :new"),
		ExpressionAttributeNames: map[string]*string{
			"#status": aws.String("status"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":new": {S: aws.String(newStatus.String())},
		},
	})
	return err
}

func UpdateDepositAsFailed(depositId, reason string) error {
	current, err := GetDepositById(depositId)
	if err != nil {
		return err
	}
	if !ValidateStatusChange(current.Status, domain.Failed) {
		return InconsistentStatusChangeError{OldStatus: current.Status, NewStatus: domain.Failed}
	}
	_, err = DdbSvc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: TableName,
		Key: map[string]*dynamodb.AttributeValue{
			"depositId": {S: aws.String(depositId)},
		},
		UpdateExpression: aws.String("SET #status = :new, failureReason = :reason"),
		ExpressionAttributeNames: map[string]*string{
			"#status": aws.String("status"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":new":    {S: aws.String(domain.Failed.String())},
			":reason": {S: aws.String(reason)},
		},
	})
	return err
}
