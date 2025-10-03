package browser

import (
	"context"
	"testing"
)

func TestNewWebDriverClient(t *testing.T) {
	client := NewWebDriverClient("http://localhost:4444")

	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.baseURL != "http://localhost:4444" {
		t.Errorf("Expected baseURL to be 'http://localhost:4444', got '%s'", client.baseURL)
	}

	if client.httpClient == nil {
		t.Fatal("Expected httpClient to be initialized")
	}

	if client.sessionID != "" {
		t.Errorf("Expected sessionID to be empty initially, got '%s'", client.sessionID)
	}
}

func TestWebDriverClientSessionManagement(t *testing.T) {
	client := NewWebDriverClient("http://localhost:4444")
	ctx := context.Background()

	// Test that deleting a session that doesn't exist just logs a warning (no error)
	err := client.DeleteSession(ctx)
	if err != nil {
		t.Errorf("Expected no error when deleting non-existent session, got: %v", err)
	}

	// Test that we can't navigate without a session
	err = client.Navigate(ctx, "https://example.com", nil)
	if err == nil {
		t.Error("Expected error when navigating without session")
	}

	// Test that we can't get URL without a session
	_, err = client.GetCurrentURL(ctx)
	if err == nil {
		t.Error("Expected error when getting URL without session")
	}

	// Test that we can't get title without a session
	_, err = client.GetTitle(ctx)
	if err == nil {
		t.Error("Expected error when getting title without session")
	}

	// Test that we can't execute script without a session
	_, err = client.ExecuteScript(ctx, "return 1", nil)
	if err == nil {
		t.Error("Expected error when executing script without session")
	}
}

func TestWebDriverClientElementOperations(t *testing.T) {
	client := NewWebDriverClient("http://localhost:4444")
	ctx := context.Background()

	// Test that we can't find elements without a session
	_, err := client.FindElement(ctx, "body")
	if err == nil {
		t.Error("Expected error when finding element without session")
	}

	// Test that we can't click elements without a session
	err = client.ClickElement(ctx, "element-id")
	if err == nil {
		t.Error("Expected error when clicking element without session")
	}

	// Test that we can't send keys without a session
	err = client.SendKeys(ctx, "element-id", "test")
	if err == nil {
		t.Error("Expected error when sending keys without session")
	}
}

func TestWebDriverClientScreenshot(t *testing.T) {
	client := NewWebDriverClient("http://localhost:4444")
	ctx := context.Background()

	// Test that we can't take screenshot without a session
	_, err := client.TakeScreenshot(ctx)
	if err == nil {
		t.Error("Expected error when taking screenshot without session")
	}
}
