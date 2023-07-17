package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAuthHandle(t *testing.T) {
	form := url.Values{}
	form.Set("username", "testuser")
	form.Set("password", "testpassword")
	form.Set("code", "123456")

	req, err := http.NewRequest("POST", "/auth", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(AuthHandle)
	handler.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v, want %v", status, http.StatusUnauthorized)
	}
}
