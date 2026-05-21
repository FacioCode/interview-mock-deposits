package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"interview_mock_deposits_go/internal/events"

	awsevents "github.com/aws/aws-lambda-go/events"
)

func TestHandleRequest_invalidSource_skipsDownstream(t *testing.T) {
	called := false
	CreateDepositFunc = func(events.PendingTransactionEvent, string) error {
		called = true
		return nil
	}
	err := HandleRequest(context.Background(), awsevents.CloudWatchEvent{
		Source:     "unrelated.source",
		DetailType: events.PendingTransactionType,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if called {
		t.Fatal("CreateDeposit should not be called for invalid source")
	}
}

func TestHandleRequest_invalidDetailType_skipsDownstream(t *testing.T) {
	called := false
	CreateDepositFunc = func(events.PendingTransactionEvent, string) error {
		called = true
		return nil
	}
	err := HandleRequest(context.Background(), awsevents.CloudWatchEvent{
		Source:     events.PendingTransactionSource,
		DetailType: "wrong-type",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if called {
		t.Fatal("CreateDeposit should not be called for invalid detail type")
	}
}

func TestHandleRequest_invalidJSON_skipsDownstream(t *testing.T) {
	called := false
	CreateDepositFunc = func(events.PendingTransactionEvent, string) error {
		called = true
		return nil
	}
	err := HandleRequest(context.Background(), awsevents.CloudWatchEvent{
		Source:     events.PendingTransactionSource,
		DetailType: events.PendingTransactionType,
		Detail:     []byte("not-json"),
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if called {
		t.Fatal("CreateDeposit should not be called for invalid JSON")
	}
}

func TestHandleRequest_validEvent_callsCreateDeposit(t *testing.T) {
	var captured events.PendingTransactionEvent
	CreateDepositFunc = func(e events.PendingTransactionEvent, requestId string) error {
		captured = e
		return nil
	}
	detail, _ := json.Marshal(events.PendingTransactionEvent{
		TransactionId: "tx-123",
		CustomerId:    "cust-456",
		Amount:        "100.00",
	})
	err := HandleRequest(context.Background(), awsevents.CloudWatchEvent{
		Source:     events.PendingTransactionSource,
		DetailType: events.PendingTransactionType,
		Detail:     detail,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.TransactionId != "tx-123" || captured.CustomerId != "cust-456" || captured.Amount != "100.00" {
		t.Fatalf("downstream received unexpected payload: %#v", captured)
	}
}

func TestHandleRequest_propagatesDownstreamError(t *testing.T) {
	CreateDepositFunc = func(events.PendingTransactionEvent, string) error {
		return errors.New("boom")
	}
	detail, _ := json.Marshal(events.PendingTransactionEvent{})
	err := HandleRequest(context.Background(), awsevents.CloudWatchEvent{
		Source:     events.PendingTransactionSource,
		DetailType: events.PendingTransactionType,
		Detail:     detail,
	})
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom error, got %v", err)
	}
}
