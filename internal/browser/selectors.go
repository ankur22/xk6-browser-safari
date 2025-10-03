package browser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// SelectorStrategy represents a selector type
type SelectorStrategy string

const (
	// Native WebDriver strategies
	StrategyCSSSelector     SelectorStrategy = "css selector"
	StrategyXPath           SelectorStrategy = "xpath"
	StrategyLinkText        SelectorStrategy = "link text"
	StrategyPartialLinkText SelectorStrategy = "partial link text"
	StrategyID              SelectorStrategy = "id"
	StrategyClassName       SelectorStrategy = "class name"
	StrategyTagName         SelectorStrategy = "tag name"

	// Custom JavaScript-based strategies
	StrategyText        SelectorStrategy = "text"
	StrategyDataTestID  SelectorStrategy = "data-testid"
	StrategyAriaLabel   SelectorStrategy = "aria-label"
	StrategyRole        SelectorStrategy = "role"
	StrategyVisibleText SelectorStrategy = "visible-text"
)

// ParsedSelector contains the parsed selector information
type ParsedSelector struct {
	Strategy SelectorStrategy
	Value    string
	IsNative bool
}

// ParseSelector parses a selector string and determines its strategy
func ParseSelector(selector string) ParsedSelector {
	// Check for explicit strategy prefixes
	if strings.HasPrefix(selector, "xpath=") {
		return ParsedSelector{StrategyXPath, strings.TrimPrefix(selector, "xpath="), true}
	}
	if strings.HasPrefix(selector, "//") || strings.HasPrefix(selector, "(//") {
		return ParsedSelector{StrategyXPath, selector, true}
	}
	if strings.HasPrefix(selector, "text=") {
		return ParsedSelector{StrategyText, strings.TrimPrefix(selector, "text="), false}
	}
	if strings.HasPrefix(selector, "visible-text=") {
		return ParsedSelector{StrategyVisibleText, strings.TrimPrefix(selector, "visible-text="), false}
	}
	if strings.HasPrefix(selector, "id=") {
		return ParsedSelector{StrategyID, strings.TrimPrefix(selector, "id="), true}
	}
	if strings.HasPrefix(selector, "class=") {
		return ParsedSelector{StrategyClassName, strings.TrimPrefix(selector, "class="), true}
	}
	if strings.HasPrefix(selector, "tag=") {
		return ParsedSelector{StrategyTagName, strings.TrimPrefix(selector, "tag="), true}
	}
	if strings.HasPrefix(selector, "link=") {
		return ParsedSelector{StrategyLinkText, strings.TrimPrefix(selector, "link="), true}
	}
	if strings.HasPrefix(selector, "partial-link=") {
		return ParsedSelector{StrategyPartialLinkText, strings.TrimPrefix(selector, "partial-link="), true}
	}
	if strings.HasPrefix(selector, "data-testid=") {
		return ParsedSelector{StrategyDataTestID, strings.TrimPrefix(selector, "data-testid="), false}
	}
	if strings.HasPrefix(selector, "aria-label=") {
		return ParsedSelector{StrategyAriaLabel, strings.TrimPrefix(selector, "aria-label="), false}
	}
	if strings.HasPrefix(selector, "role=") {
		return ParsedSelector{StrategyRole, strings.TrimPrefix(selector, "role="), false}
	}

	// Default to CSS selector
	return ParsedSelector{StrategyCSSSelector, selector, true}
}

// FindElementWithStrategy finds an element using the parsed selector strategy
func (c *WebDriverClient) FindElementWithStrategy(ctx context.Context, selector string) (string, error) {
	parsed := ParseSelector(selector)

	if parsed.IsNative {
		return c.findElementNative(ctx, string(parsed.Strategy), parsed.Value)
	}

	return c.findElementCustom(ctx, parsed.Strategy, parsed.Value)
}

// findElementNative uses WebDriver's native element finding
func (c *WebDriverClient) findElementNative(ctx context.Context, strategy, value string) (string, error) {
	if c.sessionID == "" {
		return "", fmt.Errorf("no active session")
	}

	payload := map[string]string{"using": strategy, "value": value}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal find element payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/session/"+c.sessionID+"/element", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create find element request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to find element: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Try to get error details from response
		var errorBody map[string]interface{}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&errorBody); decodeErr == nil {
			if value, ok := errorBody["value"].(map[string]interface{}); ok {
				if message, ok := value["message"].(string); ok {
					return "", fmt.Errorf("find element failed with status %d: strategy=%s, selector=%s, error=%s", resp.StatusCode, strategy, value, message)
				}
			}
		}
		return "", fmt.Errorf("find element failed with status %d: strategy=%s, selector=%s", resp.StatusCode, strategy, value)
	}

	var elementResp struct {
		Value struct {
			ElementID string `json:"element-6066-11e4-a52e-4f735466cecf"`
			ELEMENT   string `json:"ELEMENT"` // Fallback for older WebDriver
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&elementResp); err != nil {
		return "", fmt.Errorf("failed to decode element response: %w", err)
	}

	// Try W3C standard field first, fallback to older ELEMENT field
	if elementResp.Value.ElementID != "" {
		return elementResp.Value.ElementID, nil
	}
	if elementResp.Value.ELEMENT != "" {
		return elementResp.Value.ELEMENT, nil
	}

	return "", fmt.Errorf("element not found")
}

// findElementCustom uses JavaScript to find elements with custom strategies
func (c *WebDriverClient) findElementCustom(ctx context.Context, strategy SelectorStrategy, value string) (string, error) {
	script := generateSelectorScript(strategy, value)

	result, err := c.ExecuteScript(ctx, script, nil)
	if err != nil {
		return "", fmt.Errorf("failed to execute selector script: %w", err)
	}

	// Check if element was found
	if result == nil {
		return "", fmt.Errorf("element not found")
	}

	fmt.Println("Found element:", result)

	// WebDriver returns element references as maps
	if elemMap, ok := result.(map[string]interface{}); ok {
		// Try W3C standard key
		if elemID, ok := elemMap["element-6066-11e4-a52e-4f735466cecf"].(string); ok {
			return elemID, nil
		}
		// Try older ELEMENT key
		if elemID, ok := elemMap["ELEMENT"].(string); ok {
			return elemID, nil
		}
	}

	return "", fmt.Errorf("invalid element reference returned")
}

// generateSelectorScript generates JavaScript code for custom selector strategies
func generateSelectorScript(strategy SelectorStrategy, value string) string {
	escapedValue := strings.ReplaceAll(value, `"`, `\"`)

	switch strategy {
	case StrategyText:
		return fmt.Sprintf(`
			// Find the most specific (deepest) element with exact matching text
			var elements = Array.from(document.querySelectorAll('*'));
			var matches = elements.filter(function(el) {
				// Get only the direct text content (not from children)
				var directText = Array.from(el.childNodes)
					.filter(function(node) { return node.nodeType === 3; })
					.map(function(node) { return node.textContent; })
					.join('').trim();
				return directText === "%s" || el.textContent.trim() === "%s";
			});
			// Return the deepest (most specific) match
			if (matches.length > 0) {
				return matches[matches.length - 1];
			}
			return null;
		`, escapedValue, escapedValue)

	case StrategyVisibleText:
		return fmt.Sprintf(`
			// Find the most specific visible element containing the text
			var elements = Array.from(document.querySelectorAll('*'));
			var matches = elements.filter(function(el) {
				// Check visibility
				if (el.offsetWidth === 0 || el.offsetHeight === 0) return false;
				var style = window.getComputedStyle(el);
				if (style.display === 'none' || style.visibility === 'hidden') return false;
				
				// Check text content
				var text = el.textContent ? el.textContent.trim() : '';
				return text.includes("%s");
			});
			
			// Return the smallest (most specific) element
			// Sort by total descendants count (fewer = more specific)
			matches.sort(function(a, b) {
				return a.getElementsByTagName('*').length - b.getElementsByTagName('*').length;
			});
			
			return matches.length > 0 ? matches[0] : null;
		`, escapedValue)

	case StrategyDataTestID:
		return fmt.Sprintf(`return document.querySelector('[data-testid="%s"]');`, escapedValue)

	case StrategyAriaLabel:
		return fmt.Sprintf(`return document.querySelector('[aria-label="%s"]');`, escapedValue)

	case StrategyRole:
		return fmt.Sprintf(`return document.querySelector('[role="%s"]');`, escapedValue)

	default:
		// Fallback to CSS selector
		return fmt.Sprintf(`return document.querySelector("%s");`, escapedValue)
	}
}

// generateAllSelectorScript generates JavaScript code to find ALL elements (not just one)
func generateAllSelectorScript(strategy SelectorStrategy, value string) string {
	escapedValue := strings.ReplaceAll(value, `"`, `\"`)

	switch strategy {
	case StrategyText:
		return fmt.Sprintf(`
			var elements = Array.from(document.querySelectorAll('*'));
			return elements.filter(function(el) {
				var directText = Array.from(el.childNodes)
					.filter(function(node) { return node.nodeType === 3; })
					.map(function(node) { return node.textContent; })
					.join('').trim();
				return directText === "%s" || el.textContent.trim() === "%s";
			});
		`, escapedValue, escapedValue)

	case StrategyVisibleText:
		return fmt.Sprintf(`
			var elements = Array.from(document.querySelectorAll('*'));
			return elements.filter(function(el) {
				if (el.offsetWidth === 0 || el.offsetHeight === 0) return false;
				var style = window.getComputedStyle(el);
				if (style.display === 'none' || style.visibility === 'hidden') return false;
				var text = el.textContent ? el.textContent.trim() : '';
				return text.includes("%s");
			});
		`, escapedValue)

	case StrategyDataTestID:
		return fmt.Sprintf(`return Array.from(document.querySelectorAll('[data-testid="%s"]'));`, escapedValue)

	case StrategyAriaLabel:
		return fmt.Sprintf(`return Array.from(document.querySelectorAll('[aria-label="%s"]'));`, escapedValue)

	case StrategyRole:
		return fmt.Sprintf(`return Array.from(document.querySelectorAll('[role="%s"]'));`, escapedValue)

	default:
		// Fallback to CSS selector for all
		return fmt.Sprintf(`return Array.from(document.querySelectorAll("%s"));`, escapedValue)
	}
}

// IsRegex checks if a string is a regex pattern (enclosed in /)
func IsRegex(s string) bool {
	return len(s) >= 2 && strings.HasPrefix(s, "/") && strings.HasSuffix(s, "/")
}

// ParseRegex extracts the regex pattern from /pattern/ format
func ParseRegex(s string) (*regexp.Regexp, error) {
	if !IsRegex(s) {
		return nil, fmt.Errorf("not a regex pattern")
	}
	pattern := s[1 : len(s)-1]
	return regexp.Compile(pattern)
}
