package streamConsumer

import (
	"context"

	"interview_mock_deposits_go/deposit/src/domain"

	"github.com/aws/aws-lambda-go/events"
)

const (
	insertEvent = "INSERT"
	modifyEvent = "MODIFY"
)

func Consume(ctx context.Context, e events.DynamoDBEvent) error {
	for _, record := range e.Records {
		SendEvents(record)
	}
	return nil
}

func isNewDeposit(record events.DynamoDBEventRecord) bool {
	newItem := record.Change.NewImage
	if _, ok := newItem["status"]; !ok {
		return false
	}
	return record.EventName == insertEvent && newItem["status"].String() == domain.New.String()
}

func isDepositSent(record events.DynamoDBEventRecord) bool {
	newItem := record.Change.NewImage
	oldItem := record.Change.OldImage
	if _, ok := newItem["status"]; !ok {
		return false
	}
	return record.EventName == modifyEvent &&
		oldItem["status"].String() != newItem["status"].String() &&
		newItem["status"].String() == domain.DepositSent.String()
}
