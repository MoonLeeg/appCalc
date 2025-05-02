package orchestrator

import (
	"calculator/internal/database"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupHandlers(t *testing.T) *HTTPHandlers {
	store, err := database.NewStore(":memory:")
	if err != nil {
		if strings.Contains(err.Error(), "CGO_ENABLED") {
			t.Skipf("skip HTTP handler tests due DB init error: %v", err)
		}
		t.Fatalf("NewStore error: %v", err)
	}
	if err := store.InitDB(); err != nil {
		if strings.Contains(err.Error(), "CGO_ENABLED") {
			t.Skipf("skip HTTP handler tests due DB migration error: %v", err)
		}
		t.Fatalf("InitDB error: %v", err)
	}
	authService := NewAuthService(store, "testsecret")
	scheduler := NewScheduler(store)
	return NewHTTPHandlers(authService, store, scheduler)
}

func TestRegisterLoginCalculateFlow(t *testing.T) {
	h := setupHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/register", strings.NewReader(`{"login":"user","password":"pass123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.RegisterHandler(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Register expected %d, got %d body=%s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/register", strings.NewReader(`{"login":"user","password":"pass123"}`))
	req.Header.Set("Content-Type", "application/json")
	h.RegisterHandler(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("Duplicate register expected %d, got %d", http.StatusConflict, rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(`{"login":"user","password":"pass123"}`))
	req.Header.Set("Content-Type", "application/json")
	h.LoginHandler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Login expected %d, got %d", http.StatusOK, rec.Code)
	}
	var loginResp struct{ Token string }
	if err := json.NewDecoder(rec.Body).Decode(&loginResp); err != nil {
		t.Fatalf("Login decode error: %v", err)
	}
	if loginResp.Token == "" {
		t.Fatal("Login token empty")
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(`{"expression":"2+2"}`))
	req.Header.Set("Content-Type", "application/json")
	h.CalculateHandler(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Calculate without auth expected %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(`{"expression":"2+2"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	h.CalculateHandler(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Calculate with auth expected %d, got %d body=%s", http.StatusCreated, rec.Code, rec.Body.String())
	}
	var calcResp struct {
		Id         int64
		Expression string
		Status     string
	}
	if err := json.NewDecoder(rec.Body).Decode(&calcResp); err != nil {
		t.Fatalf("Calculate decode error: %v", err)
	}
	if calcResp.Id == 0 {
		t.Fatal("Calculate returned zero id")
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/expressions", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	h.ExpressionsHandler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expressions list expected %d, got %d", http.StatusOK, rec.Code)
	}
	var list []database.Expression
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("List decode error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("Expected 1 expression, got %d", len(list))
	}
}