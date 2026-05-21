package main

import (
	"context"
	"encoding/json"

	create_deposit "interview_mock_deposits_go/create_new_deposit/src"
	"interview_mock_deposits_go/internal/events"

	awsevents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/sirupsen/logrus"
)

var contextGetter = lambdacontext.FromContext

// CreateDepositFunc is a package-level indirection so tests can stub the
// downstream call without spinning up real services.
var CreateDepositFunc = create_deposit.CreateDeposit

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, event awsevents.CloudWatchEvent) error {
	defer func() {
		if r := recover(); r != nil {
			logrus.Panic(r)
		}
	}()

	var requestId string
	if lContext, ok := contextGetter(ctx); ok && lContext != nil {
		requestId = lContext.AwsRequestID
	}

	logrus.WithField("event", event).Info("event received")

	if event.Source != events.PendingTransactionSource {
		logrus.Error("invalid source: " + event.Source)
		return nil
	}

	if event.DetailType != events.PendingTransactionType {
		logrus.Error("invalid event type: " + event.DetailType)
		return nil
	}

	detail := events.PendingTransactionEvent{}
	if err := json.Unmarshal(event.Detail, &detail); err != nil {
		logrus.WithField("error", err.Error()).Error("invalid pending transaction detail")
		return nil
	}

	return CreateDepositFunc(detail, requestId)
}
