package partner_api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	DuplicatedTransferErrorCode = "PARTNER_DUPLICATE"
	BlockedDocumentErrorCode    = "PARTNER_BLOCKED_DOCUMENT"
)

var (
	ErrPartnerUnavailable    = errors.New("partner API unavailable")
	ErrInvalidCredentials    = errors.New("invalid partner credentials")
	ErrBlockedDocument       = errors.New("destination document is blocked")
	ErrTransferAlreadyClosed = errors.New("transfer already closed")
)

const (
	StatusConfirmed = "CONFIRMED"
)

type PartnerSecrets struct {
	ApiKey    string
	ApiSecret string
}

type TransferRequest struct {
	ID     string
	Amount float64
}

type TransferResult struct {
	ID     string
	Status string
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Package-level seams. Tests swap these to stub individual stages or
// to redirect the HTTP client at an httptest server.
var (
	GetPartnerSecrets  = getPartnerSecrets
	CreateTransferFunc = createTransfer
	CloseTransferFunc  = closeTransfer
	HttpClient         = &http.Client{Timeout: 30 * time.Second}
)

func partnerBaseURL() string {
	if u := os.Getenv("PARTNER_BASE_URL"); u != "" {
		return u
	}
	return "https://api.partner.example.com"
}

// Pay runs the two-step partner flow: create a transfer, then close it.
// Returns CONFIRMED on full success; any failure (including a failed
// close after a successful create) returns a zero-value TransferResult
// and a typed error.
func Pay(ctx context.Context, req TransferRequest) (TransferResult, error) {
	secrets, err := GetPartnerSecrets()
	if err != nil {
		return TransferResult{}, fmt.Errorf("get partner secrets: %w", err)
	}

	transferId, err := CreateTransferFunc(ctx, secrets, req)
	if err != nil {
		return TransferResult{}, fmt.Errorf("create transfer: %w", err)
	}

	if err := CloseTransferFunc(ctx, secrets, transferId); err != nil {
		return TransferResult{}, fmt.Errorf("close transfer: %w", err)
	}

	return TransferResult{ID: transferId, Status: StatusConfirmed}, nil
}

// getPartnerSecrets loads credentials from environment variables.
// In production these would come from Secrets Manager / KMS.
func getPartnerSecrets() (PartnerSecrets, error) {
	apiKey := os.Getenv("PARTNER_API_KEY")
	apiSecret := os.Getenv("PARTNER_API_SECRET")
	if apiKey == "" || apiSecret == "" {
		return PartnerSecrets{}, ErrInvalidCredentials
	}
	return PartnerSecrets{ApiKey: apiKey, ApiSecret: apiSecret}, nil
}

// createTransfer POSTs a new transfer to the partner and maps known
// error codes from the response body onto typed errors.
func createTransfer(ctx context.Context, secrets PartnerSecrets, req TransferRequest) (string, error) {
	body, err := json.Marshal(map[string]any{
		"id":     req.ID,
		"amount": req.Amount,
	})
	if err != nil {
		return "", fmt.Errorf("marshal transfer request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, partnerBaseURL()+"/transfers", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-Key", secrets.ApiKey)

	resp, err := HttpClient.Do(httpReq)
	if err != nil {
		return "", ErrPartnerUnavailable
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var out struct {
			TransferId string `json:"transferId"`
		}
		if err := json.Unmarshal(raw, &out); err != nil {
			return "", err
		}
		return out.TransferId, nil
	}

	var errBody ErrorBody
	_ = json.Unmarshal(raw, &errBody)
	switch errBody.Code {
	case DuplicatedTransferErrorCode:
		return "", fmt.Errorf("duplicated transfer: %s", errBody.Message)
	case BlockedDocumentErrorCode:
		return "", ErrBlockedDocument
	}
	return "", fmt.Errorf("create transfer status %d: %s", resp.StatusCode, errBody.Message)
}

// closeTransfer finalizes a previously-created transfer. A 409 from
// the partner is mapped to ErrTransferAlreadyClosed so the caller can
// treat it as idempotent.
func closeTransfer(ctx context.Context, secrets PartnerSecrets, transferId string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, partnerBaseURL()+"/transfers/"+transferId+"/close", nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("X-Api-Key", secrets.ApiKey)

	resp, err := HttpClient.Do(httpReq)
	if err != nil {
		return ErrPartnerUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return ErrTransferAlreadyClosed
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("close transfer status %d", resp.StatusCode)
}

// IsWebhookValid verifies an HMAC-SHA256 signature against the raw
// webhook body. The partner signs the body with apiSecret and sends
// the hex digest in the x-partner-signature header.
func IsWebhookValid(signature, body string) bool {
	if signature == "" {
		return false
	}
	secrets, err := GetPartnerSecrets()
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secrets.ApiSecret))
	mac.Write([]byte(body))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
