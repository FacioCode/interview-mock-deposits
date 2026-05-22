package main

import (
	"context"
	"encoding/json"

	deposit_in_provider "interview_mock_deposits_go/deposit_in_provider/src"
	"interview_mock_deposits_go/internal/events"

	awsevents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/sirupsen/logrus"
)

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, event awsevents.CloudWatchEvent) error {
	awsRequestId := ""
	if lc, ok := lambdacontext.FromContext(ctx); ok {
		awsRequestId = lc.AwsRequestID
	}

	defer func() {
		if r := recover(); r != nil {
			logrus.Panic(r)
		}
	}()

	logrus.WithField("event", event).Info("event received")

	if event.Source != events.Source {
		logrus.Error("invalid source: " + event.Source)
		return nil
	}

	if event.DetailType != events.DepositRequestedType {
		logrus.Error("invalid event type: " + event.DetailType)
		return nil
	}

	detail := events.DepositRequestedEvent{}
	if err := json.Unmarshal(event.Detail, &detail); err != nil {
		logrus.WithField("error", err.Error()).Error("invalid deposit requested detail")
		return nil
	}

	if err := deposit_in_provider.CreateDepositInProvider(detail, awsRequestId); err != nil {
		logrus.WithFields(logrus.Fields{"event": event, "error": err.Error()}).Error("error creating deposit in provider")
		return nil
	}

	return nil
}
