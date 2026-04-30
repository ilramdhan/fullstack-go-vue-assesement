package http_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/durianpay/fullstack-boilerplate/internal/api"
	"github.com/durianpay/fullstack-boilerplate/internal/migration"
	authHandler "github.com/durianpay/fullstack-boilerplate/internal/module/auth/handler"
	authRepo "github.com/durianpay/fullstack-boilerplate/internal/module/auth/repository"
	authUC "github.com/durianpay/fullstack-boilerplate/internal/module/auth/usecase"
	paymentHandler "github.com/durianpay/fullstack-boilerplate/internal/module/payment/handler"
	paymentRepo "github.com/durianpay/fullstack-boilerplate/internal/module/payment/repository"
	paymentUC "github.com/durianpay/fullstack-boilerplate/internal/module/payment/usecase"
	srv "github.com/durianpay/fullstack-boilerplate/internal/service/http"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"
)

func setup(t *testing.T) (http.Handler, *sql.DB) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=1")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := migration.Run(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if _, err := db.Exec(`INSERT INTO users(email, password_hash, role) VALUES (?,?,?), (?,?,?)`,
		"cs@test.com", string(hash), "cs",
		"op@test.com", string(hash), "operation",
	); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.Exec(`
INSERT INTO payments(id, merchant, amount, currency, status, created_at) VALUES
('p-comp', 'Tokopedia', 1000, 'IDR', 'completed',  ?),
('p-proc', 'Shopee',     2000, 'IDR', 'processing', ?),
('p-fail', 'Lazada',     3000, 'IDR', 'failed',     ?)`,
		now, now, now,
	); err != nil {
		t.Fatalf("seed payments: %v", err)
	}

	auc := authUC.NewAuthUsecase(authRepo.NewUserRepo(db), []byte("test-secret-test-secret"), time.Hour)
	puc := paymentUC.NewPaymentUsecase(paymentRepo.NewPaymentRepo(db))

	apiHandler := &api.APIHandler{
		Auth:    authHandler.NewAuthHandler(auc),
		Payment: paymentHandler.NewPaymentHandler(puc),
	}
	server := srv.NewServer(apiHandler, auc, "")
	return server.Routes(), db
}

func login(t *testing.T, h http.Handler, email, password string) string {
	t.Helper()
	body := strings.NewReader(`{"email":"` + email + `","password":"` + password + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/dashboard/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login %s: got %d body=%s", email, rec.Code, rec.Body.String())
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}
	return out.Token
}

func do(t *testing.T, h http.Handler, method, path, token string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealthz(t *testing.T) {
	h, _ := setup(t)
	rec := do(t, h, http.MethodGet, "/healthz", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestLogin_BadCredentials(t *testing.T) {
	h, _ := setup(t)
	rec := do(t, h, http.MethodPost, "/dashboard/v1/auth/login", "",
		strings.NewReader(`{"email":"cs@test.com","password":"WRONG"}`))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_MalformedEmailRejectedByValidator(t *testing.T) {
	h, _ := setup(t)
	rec := do(t, h, http.MethodPost, "/dashboard/v1/auth/login", "",
		strings.NewReader(`{"email":"not-an-email","password":"x"}`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPayments_RequiresToken(t *testing.T) {
	h, _ := setup(t)
	rec := do(t, h, http.MethodGet, "/dashboard/v1/payments", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestPayments_ListAndFilter(t *testing.T) {
	h, _ := setup(t)
	tok := login(t, h, "cs@test.com", "secret")

	rec := do(t, h, http.MethodGet, "/dashboard/v1/payments", tok, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Total int `json:"total"`
		Data  []struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if out.Total != 3 || len(out.Data) != 3 {
		t.Errorf("expected 3 rows, got total=%d len=%d", out.Total, len(out.Data))
	}

	rec = do(t, h, http.MethodGet, "/dashboard/v1/payments?status=processing", tok, nil)
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if out.Total != 1 || out.Data[0].Status != "processing" {
		t.Errorf("filter processing: total=%d data=%v", out.Total, out.Data)
	}

	rec = do(t, h, http.MethodGet, "/dashboard/v1/payments?status=invalid", tok, nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("invalid status: got %d", rec.Code)
	}
}

func TestPayments_Summary(t *testing.T) {
	h, _ := setup(t)
	tok := login(t, h, "cs@test.com", "secret")
	rec := do(t, h, http.MethodGet, "/dashboard/v1/payments/summary", tok, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d body=%s", rec.Code, rec.Body.String())
	}
	var s struct {
		Total      int `json:"total"`
		Completed  int `json:"completed"`
		Processing int `json:"processing"`
		Failed     int `json:"failed"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &s)
	if s.Total != 3 || s.Completed != 1 || s.Processing != 1 || s.Failed != 1 {
		t.Errorf("summary: %+v", s)
	}
}

func TestReview_CSForbidden(t *testing.T) {
	h, _ := setup(t)
	tok := login(t, h, "cs@test.com", "secret")
	rec := do(t, h, http.MethodPut, "/dashboard/v1/payments/p-proc/review", tok,
		bytes.NewReader([]byte(`{"decision":"approve"}`)))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestReview_OperationApproveProcessing(t *testing.T) {
	h, _ := setup(t)
	tok := login(t, h, "op@test.com", "secret")
	rec := do(t, h, http.MethodPut, "/dashboard/v1/payments/p-proc/review", tok,
		bytes.NewReader([]byte(`{"decision":"approve"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d body=%s", rec.Code, rec.Body.String())
	}
	var p struct {
		Status     string  `json:"status"`
		ReviewedBy *string `json:"reviewed_by"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &p)
	if p.Status != "completed" || p.ReviewedBy == nil || *p.ReviewedBy != "op@test.com" {
		t.Errorf("review result: %+v", p)
	}

	rec = do(t, h, http.MethodPut, "/dashboard/v1/payments/p-proc/review", tok,
		bytes.NewReader([]byte(`{"decision":"approve"}`)))
	if rec.Code != http.StatusConflict {
		t.Fatalf("re-review: got %d", rec.Code)
	}
}

func TestReview_OperationRejectProcessing(t *testing.T) {
	h, db := setup(t)
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.Exec(
		`INSERT INTO payments(id, merchant, amount, currency, status, created_at) VALUES (?, ?, ?, 'IDR', ?, ?)`,
		"p-proc-2", "Blibli", 500, "processing", now,
	); err != nil {
		t.Fatalf("seed: %v", err)
	}
	tok := login(t, h, "op@test.com", "secret")
	rec := do(t, h, http.MethodPut, "/dashboard/v1/payments/p-proc-2/review", tok,
		bytes.NewReader([]byte(`{"decision":"reject"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d body=%s", rec.Code, rec.Body.String())
	}
	var p struct {
		Status string `json:"status"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &p)
	if p.Status != "failed" {
		t.Errorf("expected failed, got %s", p.Status)
	}
}

func TestReview_NotFound(t *testing.T) {
	h, _ := setup(t)
	tok := login(t, h, "op@test.com", "secret")
	rec := do(t, h, http.MethodPut, "/dashboard/v1/payments/missing/review", tok,
		bytes.NewReader([]byte(`{"decision":"approve"}`)))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestReview_OperationOnCompletedConflict(t *testing.T) {
	h, _ := setup(t)
	tok := login(t, h, "op@test.com", "secret")
	rec := do(t, h, http.MethodPut, "/dashboard/v1/payments/p-comp/review", tok,
		bytes.NewReader([]byte(`{"decision":"approve"}`)))
	if rec.Code != http.StatusConflict {
		t.Fatalf("got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestReview_MissingDecisionRejectedByValidator(t *testing.T) {
	h, _ := setup(t)
	tok := login(t, h, "op@test.com", "secret")
	rec := do(t, h, http.MethodPut, "/dashboard/v1/payments/p-proc/review", tok,
		bytes.NewReader([]byte(`{}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d body=%s", rec.Code, rec.Body.String())
	}
}
