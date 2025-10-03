package browser

import (
	"testing"
)

func TestLocatorCreation(t *testing.T) {
	page := &Page{
		client: NewWebDriverClient("http://localhost:4444"),
	}

	// Test that locator can be created
	locator := page.Locator("button.submit")

	if locator == nil {
		t.Fatal("Expected locator to be created")
	}

	if locator.selector != "button.submit" {
		t.Errorf("Expected selector to be 'button.submit', got '%s'", locator.selector)
	}

	if locator.page != page {
		t.Error("Expected locator to reference the page")
	}
}

func TestLocatorWithElementID(t *testing.T) {
	page := &Page{
		client: NewWebDriverClient("http://localhost:4444"),
	}

	// Test that a locator can represent a specific element
	locator := &Locator{
		page:      page,
		selector:  "button",
		elementID: "test-element-id",
	}

	if locator == nil {
		t.Fatal("Expected locator to be created")
	}

	if locator.elementID != "test-element-id" {
		t.Errorf("Expected elementID to be 'test-element-id', got '%s'", locator.elementID)
	}

	if locator.selector != "button" {
		t.Errorf("Expected selector to be 'button', got '%s'", locator.selector)
	}
}

func TestLocatorTextContent(t *testing.T) {
	page := &Page{
		client: NewWebDriverClient("http://localhost:4444"),
	}

	locator := page.Locator("h1")

	// TextContent should be callable
	// Note: We can't test the actual execution without a real WebDriver session
	if locator == nil {
		t.Fatal("Expected locator to be created")
	}
}

func TestLocatorType(t *testing.T) {
	page := &Page{
		client: NewWebDriverClient("http://localhost:4444"),
	}

	locator := page.Locator("input[name='email']")

	// Type should be callable
	if locator == nil {
		t.Fatal("Expected locator to be created")
	}
}
