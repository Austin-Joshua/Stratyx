package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

type loginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type item struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type itemListResponse struct {
	Items []item `json:"items"`
	Total int    `json:"total"`
}

func TestAPIAuthAndItemsCRUD(t *testing.T) {
	if os.Getenv("RUN_API_TESTS") != "1" {
		t.Skip("set RUN_API_TESTS=1 to run integration API tests")
	}

	baseURL := os.Getenv("TEST_API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	email := fmt.Sprintf("test-%d@stratyx.test", time.Now().UnixNano())
	password := "Password123!"

	registerPayload := map[string]string{
		"name":     "Integration User",
		"email":    email,
		"password": password,
	}
	doJSON(t, http.MethodPost, baseURL+"/api/auth/register", "", registerPayload, nil)

	var login loginResponse
	doJSON(t, http.MethodPost, baseURL+"/api/auth/login", "", map[string]string{
		"email":    email,
		"password": password,
	}, &login)
	if login.AccessToken == "" {
		t.Fatal("expected non-empty access token")
	}

	var created item
	doJSON(t, http.MethodPost, baseURL+"/api/items", login.AccessToken, map[string]string{
		"title":       "Integration Item",
		"description": "Created by API test",
	}, &created)
	if created.ID == "" {
		t.Fatal("expected created item id")
	}

	var list itemListResponse
	doJSON(t, http.MethodGet, baseURL+"/api/items?page=1&pageSize=10&search=integration&sortBy=updatedAt&sortOrder=desc", login.AccessToken, nil, &list)
	if list.Total < 1 {
		t.Fatalf("expected at least one item, got total=%d", list.Total)
	}

	var updated item
	doJSON(t, http.MethodPut, baseURL+"/api/items/"+created.ID, login.AccessToken, map[string]string{
		"title":       "Updated Integration Item",
		"description": "Updated by API test",
	}, &updated)
	if updated.Title != "Updated Integration Item" {
		t.Fatalf("expected updated title, got %q", updated.Title)
	}

	doJSON(t, http.MethodDelete, baseURL+"/api/items/"+created.ID, login.AccessToken, nil, nil)
}

func doJSON(t *testing.T, method, url, accessToken string, payload interface{}, out interface{}) {
	t.Helper()
	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		t.Fatalf("request failed: %s %s -> %d", method, url, resp.StatusCode)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("decode response: %v", err)
		}
	}
}
