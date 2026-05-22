package main

import (
	"context"

	streamConsumer "interview_mock_deposits_go/stream_consumer/src"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
)

func HandleRequest(ctx context.Context, e events.DynamoDBEvent) error {
	defer func() {
		if r := recover(); r != nil {
			logrus.Panic(r)
		}
	}()
	logrus.WithField("recordCount", len(e.Records)).Info("stream consumer starts")
	return streamConsumer.Consume(ctx, e)
}

func main() {
	lambda.Start(HandleRequest)
}
