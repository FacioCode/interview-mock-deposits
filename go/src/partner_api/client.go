package partner_api

import (
	"context"
	"errors"
)

var ErrPartnerUnavailable = errors.New("partner API unavailable")

type TransferRequest struct {
	ID     string
	Amount float64
}

type TransferResult struct {
	ID     string
	Status string
}

func Pay(ctx context.Context, req TransferRequest) (TransferResult, error) {
	// TODO: replace with real integration
	return TransferResult{}, ErrPartnerUnavailable
}

func IsWebhookValid(signature, body string) bool {
	// TODO: replace with real signature verification
	return signature != ""
}
