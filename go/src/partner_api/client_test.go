package partner_api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// Helpers ---------------------------------------------------------------

func withEnv(t *testing.T, key, value string) {
	t.Helper()
	prev, had := os.LookupEnv(key)
	t.Setenv(key, value)
	t.Cleanup(func() {
		if had {
			os.Setenv(key, prev)
		} else {
			os.Unsetenv(key)
		}
	})
}

func resetSeams(t *testing.T) {
	t.Helper()
	origSecrets := GetPartnerSecrets
	origCreate := CreateTransferFunc
	origClose := CloseTransferFunc
	t.Cleanup(func() {
		GetPartnerSecrets = origSecrets
		CreateTransferFunc = origCreate
		CloseTransferFunc = origClose
	})
}

// Pay orchestration -----------------------------------------------------

func TestPay_secretsFailure_propagates(t *testing.T) {
	resetSeams(t)
	GetPartnerSecrets = func() (PartnerSecrets, error) {
		return PartnerSecrets{}, ErrInvalidCredentials
	}

	_, err := Pay(context.Background(), TransferRequest{ID: "d-1", Amount: 10})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials wrapped, got %v", err)
	}
}

func TestPay_createFailure_skipsCloseAndPropagates(t *testing.T) {
	resetSeams(t)
	GetPartnerSecrets = func() (PartnerSecrets, error) {
		return PartnerSecrets{ApiKey: "k", ApiSecret: "s"}, nil
	}
	CreateTransferFunc = func(ctx context.Context, _ PartnerSecrets, _ TransferRequest) (string, error) {
		return "", ErrBlockedDocument
	}
	closeCalled := false
	CloseTransferFunc = func(ctx context.Context, _ PartnerSecrets, _ string) error {
		closeCalled = true
		return nil
	}

	_, err := Pay(context.Background(), TransferRequest{ID: "d-1", Amount: 10})
	if !errors.Is(err, ErrBlockedDocument) {
		t.Fatalf("want ErrBlockedDocument wrapped, got %v", err)
	}
	if closeCalled {
		t.Fatal("close should not be called when create fails")
	}
}

func TestPay_closeFailure_returnsZeroResultAndError(t *testing.T) {
	resetSeams(t)
	GetPartnerSecrets = func() (PartnerSecrets, error) {
		return PartnerSecrets{ApiKey: "k", ApiSecret: "s"}, nil
	}
	CreateTransferFunc = func(ctx context.Context, _ PartnerSecrets, _ TransferRequest) (string, error) {
		return "tx-123", nil
	}
	CloseTransferFunc = func(ctx context.Context, _ PartnerSecrets, _ string) error {
		return ErrPartnerUnavailable
	}

	result, err := Pay(context.Background(), TransferRequest{ID: "d-1", Amount: 10})
	if !errors.Is(err, ErrPartnerUnavailable) {
		t.Fatalf("want ErrPartnerUnavailable wrapped, got %v", err)
	}
	if result != (TransferResult{}) {
		t.Fatalf("want zero-value TransferResult, got %#v", result)
	}
}

func TestPay_success_returnsConfirmed(t *testing.T) {
	resetSeams(t)
	GetPartnerSecrets = func() (PartnerSecrets, error) {
		return PartnerSecrets{ApiKey: "k", ApiSecret: "s"}, nil
	}
	CreateTransferFunc = func(ctx context.Context, _ PartnerSecrets, _ TransferRequest) (string, error) {
		return "tx-123", nil
	}
	CloseTransferFunc = func(ctx context.Context, _ PartnerSecrets, _ string) error {
		return nil
	}

	result, err := Pay(context.Background(), TransferRequest{ID: "d-1", Amount: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "tx-123" || result.Status != StatusConfirmed {
		t.Fatalf("want {tx-123 CONFIRMED}, got %#v", result)
	}
}

// HTTP-level tests (createTransfer / closeTransfer) ---------------------

func TestCreateTransfer_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transfers" || r.Method != http.MethodPost {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("X-Api-Key"); got != "k" {
			t.Errorf("X-Api-Key header: want k, got %q", got)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"transferId":"tx-abc"}`))
	}))
	t.Cleanup(srv.Close)
	withEnv(t, "PARTNER_BASE_URL", srv.URL)

	id, err := createTransfer(context.Background(),
		PartnerSecrets{ApiKey: "k", ApiSecret: "s"},
		TransferRequest{ID: "d-1", Amount: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "tx-abc" {
		t.Fatalf("want tx-abc, got %q", id)
	}
}

func TestCreateTransfer_mapsDuplicatedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":"PARTNER_DUPLICATE","message":"already exists"}`))
	}))
	t.Cleanup(srv.Close)
	withEnv(t, "PARTNER_BASE_URL", srv.URL)

	_, err := createTransfer(context.Background(),
		PartnerSecrets{ApiKey: "k", ApiSecret: "s"},
		TransferRequest{ID: "d-1", Amount: 10})
	if err == nil || !contains(err.Error(), "duplicated transfer") {
		t.Fatalf("want duplicated-transfer error, got %v", err)
	}
}

func TestCreateTransfer_mapsBlockedDocument(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":"PARTNER_BLOCKED_DOCUMENT","message":"blocked"}`))
	}))
	t.Cleanup(srv.Close)
	withEnv(t, "PARTNER_BASE_URL", srv.URL)

	_, err := createTransfer(context.Background(),
		PartnerSecrets{ApiKey: "k", ApiSecret: "s"},
		TransferRequest{ID: "d-1", Amount: 10})
	if !errors.Is(err, ErrBlockedDocument) {
		t.Fatalf("want ErrBlockedDocument, got %v", err)
	}
}

func TestCreateTransfer_networkError_returnsUnavailable(t *testing.T) {
	withEnv(t, "PARTNER_BASE_URL", "http://127.0.0.1:1")

	_, err := createTransfer(context.Background(),
		PartnerSecrets{ApiKey: "k", ApiSecret: "s"},
		TransferRequest{ID: "d-1", Amount: 10})
	if !errors.Is(err, ErrPartnerUnavailable) {
		t.Fatalf("want ErrPartnerUnavailable, got %v", err)
	}
}

func TestCloseTransfer_conflict_returnsAlreadyClosed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	t.Cleanup(srv.Close)
	withEnv(t, "PARTNER_BASE_URL", srv.URL)

	err := closeTransfer(context.Background(),
		PartnerSecrets{ApiKey: "k", ApiSecret: "s"}, "tx-1")
	if !errors.Is(err, ErrTransferAlreadyClosed) {
		t.Fatalf("want ErrTransferAlreadyClosed, got %v", err)
	}
}

// Webhook signature -----------------------------------------------------

func TestIsWebhookValid_validHmac(t *testing.T) {
	withEnv(t, "PARTNER_API_KEY", "key")
	withEnv(t, "PARTNER_API_SECRET", "topsecret")

	body := `{"status":"DONE"}`
	mac := hmac.New(sha256.New, []byte("topsecret"))
	mac.Write([]byte(body))
	sig := hex.EncodeToString(mac.Sum(nil))

	if !IsWebhookValid(sig, body) {
		t.Fatal("valid signature should pass")
	}
}

func TestIsWebhookValid_emptySignature(t *testing.T) {
	withEnv(t, "PARTNER_API_KEY", "key")
	withEnv(t, "PARTNER_API_SECRET", "topsecret")

	if IsWebhookValid("", `{"status":"DONE"}`) {
		t.Fatal("empty signature should fail")
	}
}

func TestIsWebhookValid_mismatched(t *testing.T) {
	withEnv(t, "PARTNER_API_KEY", "key")
	withEnv(t, "PARTNER_API_SECRET", "topsecret")

	if IsWebhookValid("deadbeef", `{"status":"DONE"}`) {
		t.Fatal("mismatched signature should fail")
	}
}

func TestIsWebhookValid_noCredentials(t *testing.T) {
	os.Unsetenv("PARTNER_API_KEY")
	os.Unsetenv("PARTNER_API_SECRET")

	if IsWebhookValid("anything", "body") {
		t.Fatal("missing credentials should fail")
	}
}

// strings.Contains without the import noise for these tiny checks.
func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
