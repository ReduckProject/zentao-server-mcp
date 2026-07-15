package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestGetConnectionStatusDoesNotExposeSecrets(t *testing.T) {
	tm := NewTokenManager()
	tm.config = &Config{
		BaseURL:     "https://zentao.example.test/api.php/v1",
		Account:     "tester",
		Password:    "should-not-be-returned",
		TokenExpiry: 86400,
	}
	tm.cache = &TokenCache{
		Token:      "should-not-be-returned",
		ExpireTime: time.Now().Add(time.Hour),
	}

	status := tm.GetConnectionStatus()
	if status["configured"] != true || status["has_token"] != true || status["expired"] != false {
		t.Fatalf("GetConnectionStatus() = %#v", status)
	}
	if _, exists := status["token"]; exists {
		t.Fatalf("GetConnectionStatus() exposed token: %#v", status)
	}
	if _, exists := status["password"]; exists {
		t.Fatalf("GetConnectionStatus() exposed password: %#v", status)
	}
}

func TestGetProfileHandlerIncludesConnectionStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api.php/v1/tokens", func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(w, map[string]interface{}{"token": "internal-token"})
	})
	mux.HandleFunc("/api.php/v1/user", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Token") != "internal-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		writeTestJSON(w, map[string]interface{}{
			"profile": map[string]interface{}{"id": 1, "account": "tester", "realname": "Test User"},
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	oldTokenManager := globalTokenManager
	globalTokenManager = NewTokenManager()
	globalTokenManager.config = &Config{
		BaseURL:     server.URL + "/api.php/v1",
		Account:     "tester",
		Password:    "secret",
		TokenExpiry: 86400,
	}
	globalTokenManager.client = NewZentaoClient(globalTokenManager.config.BaseURL)
	defer func() { globalTokenManager = oldTokenManager }()

	result, err := getProfileHandler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("getProfileHandler() error = %v", err)
	}
	if result.IsError || len(result.Content) != 1 {
		t.Fatalf("getProfileHandler() result = %#v", result)
	}
	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("getProfileHandler() content type = %T", result.Content[0])
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(text.Text), &response); err != nil {
		t.Fatalf("parse get_profile response: %v", err)
	}
	profile, ok := response["profile"].(map[string]interface{})
	if !ok || profile["account"] != "tester" {
		t.Fatalf("get_profile profile = %#v", response["profile"])
	}
	status, ok := response["connection_status"].(map[string]interface{})
	if !ok || status["connected"] != true || status["configured"] != true {
		t.Fatalf("get_profile connection_status = %#v", response["connection_status"])
	}
	if _, exists := status["token"]; exists {
		t.Fatalf("get_profile exposed token: %#v", status)
	}
}
