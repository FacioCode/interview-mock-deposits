package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	deposit "interview_mock_deposits_go/deposit/src"
	"interview_mock_deposits_go/internal/events"
	"interview_mock_deposits_go/internal/testUtils"

	awsevents "github.com/aws/aws-lambda-go/events"
)

// Integration tests for the create_new_deposit lambda: handler -> create_deposit
// -> deposit.CreateDepositWithUserData -> DynamoDB. Run via:
//
//   ./with-dynamodb-local.sh go test ./...
//
// The wrapper script starts DynamoDB Local, creates the table, and exports
// AWS_ENDPOINT_URL_DYNAMODB so the SDK points at it. Tests skip with a clear
// message when that env var isn't set.

func skipIfNoDynamoLocal(t *testing.T) {
	t.Helper()
	if os.Getenv("AWS_ENDPOINT_URL_DYNAMODB") == "" {
		t.Skip("DynamoDB Local not detected; run via `./with-dynamodb-local.sh go test ./...`")
	}
}

func makeEventDetail(t *testing.T, customerId, transactionId string) []byte {
	t.Helper()
	d, err := json.Marshal(events.PendingTransactionEvent{
		TransactionId: transactionId,
		CustomerId:    customerId,
		Amount:        "100.00",
		Name:          "Test User",
		Document:      "00000000000",
		BankAccount: events.BankAccount{
			Bank:    "001",
			Branch:  "0001",
			Account: "12345",
		},
	})
	if err != nil {
		t.Fatalf("failed to marshal event detail: %v", err)
	}
	return d
}

func TestHandleRequest_invalidSource_persistsNothing(t *testing.T) {
	skipIfNoDynamoLocal(t)
	customerId := testUtils.GenerateRandomId()

	err := HandleRequest(context.Background(), awsevents.CloudWatchEvent{
		Source:     "unrelated.source",
		DetailType: events.PendingTransactionType,
		Detail:     makeEventDetail(t, customerId, testUtils.GenerateRandomId()),
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	saved, err := deposit.GetDepositByUserId(customerId)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(saved) != 0 {
		t.Fatalf("expected 0 deposits, got %d", len(saved))
	}
}

func TestHandleRequest_invalidDetailType_persistsNothing(t *testing.T) {
	skipIfNoDynamoLocal(t)
	customerId := testUtils.GenerateRandomId()

	err := HandleRequest(context.Background(), awsevents.CloudWatchEvent{
		Source:     events.PendingTransactionSource,
		DetailType: "wrong-type",
		Detail:     makeEventDetail(t, customerId, testUtils.GenerateRandomId()),
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	saved, err := deposit.GetDepositByUserId(customerId)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(saved) != 0 {
		t.Fatalf("expected 0 deposits, got %d", len(saved))
	}
}

func TestHandleRequest_invalidJSON_persistsNothing(t *testing.T) {
	skipIfNoDynamoLocal(t)
	customerId := testUtils.GenerateRandomId()

	err := HandleRequest(context.Background(), awsevents.CloudWatchEvent{
		Source:     events.PendingTransactionSource,
		DetailType: events.PendingTransactionType,
		Detail:     []byte("not-json"),
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	saved, err := deposit.GetDepositByUserId(customerId)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(saved) != 0 {
		t.Fatalf("expected 0 deposits, got %d", len(saved))
	}
}

func TestHandleRequest_validEvent_persistsDeposit(t *testing.T) {
	skipIfNoDynamoLocal(t)
	customerId := testUtils.GenerateRandomId()
	transactionId := testUtils.GenerateRandomId()

	err := HandleRequest(context.Background(), awsevents.CloudWatchEvent{
		Source:     events.PendingTransactionSource,
		DetailType: events.PendingTransactionType,
		Detail:     makeEventDetail(t, customerId, transactionId),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	saved, err := deposit.GetDepositByUserId(customerId)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(saved) != 1 {
		t.Fatalf("expected 1 deposit, got %d", len(saved))
	}
	d := saved[0]
	if d.UserId != customerId {
		t.Errorf("UserId: want %q, got %q", customerId, d.UserId)
	}
	if d.TransactionId != transactionId {
		t.Errorf("TransactionId: want %q, got %q", transactionId, d.TransactionId)
	}
	if d.Amount != 100.0 {
		t.Errorf("Amount: want 100.0, got %v", d.Amount)
	}
	if d.UserData == nil || d.UserData.Document != "00000000000" {
		t.Errorf("UserData.Document missing or wrong: %#v", d.UserData)
	}
	if d.UserData.BankAccount.Bank != "001" {
		t.Errorf("BankAccount.Bank: want 001, got %q", d.UserData.BankAccount.Bank)
	}
}
