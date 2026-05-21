package main

import (
	"context"
	"encoding/json"

	deposit "interview_mock_deposits_go/deposit/src"
	"interview_mock_deposits_go/partner_api"
	webhook "interview_mock_deposits_go/webhook/src"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
)

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Panic(r)
		}
	}()

	logrus.WithField("request", request).Info("webhook request received")

	signature, ok := request.Headers["x-partner-signature"]
	if !ok || !partner_api.IsWebhookValid(signature, request.Body) {
		logrus.Warn("invalid webhook request")
		return events.APIGatewayProxyResponse{StatusCode: 403}, nil
	}

	var body webhook.Body
	if err := json.Unmarshal([]byte(request.Body), &body); err != nil {
		logrus.Error(err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	if err := webhook.HandleWebhook(body); err != nil {
		return handleWebhookError(err, body), nil
	}
	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func handleWebhookError(err error, body webhook.Body) events.APIGatewayProxyResponse {
	logFields := logrus.Fields{
		"depositId":             body.Data.IntegrationId,
		"depositProviderStatus": body.Data.Status,
	}
	if _, ok := err.(deposit.InconsistentStatusChangeError); ok {
		logrus.WithFields(logFields).Log(logrus.FatalLevel, err)
		return events.APIGatewayProxyResponse{StatusCode: 200}
	}
	logrus.WithFields(logFields).Error(err)
	return events.APIGatewayProxyResponse{StatusCode: 500}
}
