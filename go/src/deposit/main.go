package main

import (
	"context"
	"encoding/json"
	"errors"
	deposit "interview_mock_deposits_go/deposit/src"
	"interview_mock_deposits_go/deposit/src/dto"
	"interview_mock_deposits_go/utils/api"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
)

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Panic(r)
		}
	}()

	if request.HTTPMethod != "GET" {
		return api.ErrorResponse(errors.New("invalid method"), http.StatusNotImplemented)
	}

	if depositId, ok := request.PathParameters["depositId"]; ok {
		return handleGetById(depositId)
	}

	if userId, ok := request.QueryStringParameters["userId"]; ok {
		return handleListByUserId(userId)
	}

	return api.WarningResponse(errors.New("userId missing"), http.StatusBadRequest)
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	lambda.Start(HandleRequest)
}

func handleListByUserId(userId string) (events.APIGatewayProxyResponse, error) {
	logrus.WithField("userId", userId).Info("new request to get deposit list")

	userDeposits, err := deposit.GetDepositByUserId(userId)
	if err != nil {
		return api.ErrorResponse(err, http.StatusInternalServerError)
	}

	var formatted []dto.FormattedDeposit
	for _, d := range userDeposits {
		formatted = append(formatted, dto.FormatResponse(d))
	}

	body, err := json.Marshal(formatted)
	if err != nil {
		return api.ErrorResponse(err, http.StatusInternalServerError)
	}

	return api.Response(string(body), http.StatusOK)
}

func handleGetById(depositId string) (events.APIGatewayProxyResponse, error) {
	logrus.WithField("depositId", depositId).Info("new request to get deposit by id")
	depositData, err := deposit.GetDepositById(depositId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return api.Response("", http.StatusNotFound)
		}
		return api.ErrorResponse(err, http.StatusInternalServerError)
	}

	body, err := json.Marshal(dto.FormatResponse(depositData))
	if err != nil {
		return api.ErrorResponse(err, http.StatusInternalServerError)
	}

	return api.Response(string(body), http.StatusOK)
}
