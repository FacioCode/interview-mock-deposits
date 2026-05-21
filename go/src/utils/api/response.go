package api

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/sirupsen/logrus"
)

func buildResponse(body string, statusCode int) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       body,
	}
}

func buildMessageBody(message string) string {
	return fmt.Sprintf(`{"message": "%s"}`, message)
}

func Response(body string, statusCode int) (events.APIGatewayProxyResponse, error) {
	logrus.WithFields(logrus.Fields{"body": body, "statusCode": statusCode}).Info("API response")
	return buildResponse(body, statusCode), nil
}

func WarningResponse(err error, statusCode int) (events.APIGatewayProxyResponse, error) {
	body := buildMessageBody(err.Error())
	logrus.WithFields(logrus.Fields{"body": body, "statusCode": statusCode}).Warn("API response")
	return buildResponse(body, statusCode), err
}

func ErrorResponse(err error, statusCode int) (events.APIGatewayProxyResponse, error) {
	body := buildMessageBody(err.Error())
	logrus.WithFields(logrus.Fields{"body": body, "statusCode": statusCode}).Error("API response")
	return buildResponse(body, statusCode), err
}
