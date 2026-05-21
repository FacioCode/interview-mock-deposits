package partner_api

import (
	"context"
	"errors"
	"testing"
)

func TestPay_returnsUnavailable(t *testing.T) {
	_, err := Pay(context.Background(), TransferRequest{})
	if !errors.Is(err, ErrPartnerUnavailable) {
		t.Fatalf("expected ErrPartnerUnavailable, got %v", err)
	}
}

func TestIsWebhookValid(t *testing.T) {
	if IsWebhookValid("", "body") {
		t.Fatal("empty signature should be invalid")
	}
	if !IsWebhookValid("sig", "body") {
		t.Fatal("non-empty signature should be valid in stub")
	}
}
