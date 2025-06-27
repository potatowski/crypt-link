package response

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Helper to decode JSON from response body
func decodeJSONBody(t *testing.T, body *bytes.Buffer, v interface{}) {
	t.Helper()
	err := json.NewDecoder(body).Decode(v)
	if err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
}

func TestJSON_WithData(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]string{"foo": "bar"}

	JSON(rr, http.StatusCreated, data)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	var resp map[string]string
	decodeJSONBody(t, rr.Body, &resp)
	if resp["foo"] != "bar" {
		t.Errorf("expected foo=bar, got foo=%s", resp["foo"])
	}
}

func TestJSON_NilData(t *testing.T) {
	rr := httptest.NewRecorder()

	JSON(rr, http.StatusNoContent, nil)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	if rr.Body.Len() != 0 {
		t.Errorf("expected empty body, got %q", rr.Body.String())
	}
}
