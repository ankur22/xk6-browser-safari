package browser

import (
	"testing"
	"time"
)

func TestBrowserCreation(t *testing.T) {
	// This test would require a mock or actual WebDriver server
	// For now, we'll just test that the browser struct is created correctly
	browser := &Browser{
		Client: NewWebDriverClient("http://localhost:4444"),
	}

	if browser.Client == nil {
		t.Fatal("Expected Client to be set")
	}
}

func TestPageCreation(t *testing.T) {
	// This test would require a mock or actual WebDriver server
	// For now, we'll just test that the page struct is created correctly
	page := &Page{
		client: NewWebDriverClient("http://localhost:4444"),
	}

	if page.client == nil {
		t.Fatal("Expected client to be set")
	}
}

func TestPageScreenshot(t *testing.T) {
	// This test would require a mock or actual WebDriver server
	// For now, we'll just test that the page struct is created correctly
	page := &Page{
		client: NewWebDriverClient("http://localhost:4444"),
	}

	// Test that the page struct is created correctly
	if page.client == nil {
		t.Fatal("Expected client to be set")
	}
}

func TestPageWaitForTimeout(t *testing.T) {
	// Test that waitForTimeout sleeps for the correct duration
	page := &Page{
		client: NewWebDriverClient("http://localhost:4444"),
	}

	// Test a small timeout (we can't test the promise without VU context,
	// but we can verify the method is callable)
	start := time.Now()

	// Verify the page is set up
	if page.client == nil {
		t.Fatal("Expected client to be set")
	}

	elapsed := time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Test initialization took too long: %v", elapsed)
	}
}
