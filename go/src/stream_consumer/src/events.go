package streamConsumer

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/sirupsen/logrus"
)

func SendEvents(record events.DynamoDBEventRecord) {
	switch {
	case isNewDeposit(record):
		sendEvent("deposit-requested", record.Change.NewImage)
	case isDepositSent(record):
		sendEvent("deposit-sent", record.Change.NewImage)
	}
}

func sendEvent(detailType string, item map[string]events.DynamoDBAttributeValue) {
	// TODO: replace with real event publication (EventBridge PutEvents)
	logrus.WithFields(logrus.Fields{
		"detailType": detailType,
		"depositId":  item["depositId"].String(),
	}).Info("event would be published")
}
