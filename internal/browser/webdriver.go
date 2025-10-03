package browser

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"strings"
	"time"
)

// WebDriverClient handles communication with Safari WebDriver
type WebDriverClient struct {
	baseURL    string
	httpClient *http.Client
	sessionID  string
}

// WebDriverSession represents a WebDriver session
type WebDriverSession struct {
	SessionID    string                 `json:"sessionId"`
	Capabilities map[string]interface{} `json:"capabilities"`
}

// WebDriverResponse represents a standard WebDriver response
type WebDriverResponse struct {
	Value     interface{} `json:"value"`
	SessionID string      `json:"sessionId,omitempty"`
}

// NewWebDriverClient creates a new WebDriver client for Safari
func NewWebDriverClient(baseURL string) *WebDriverClient {
	return &WebDriverClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAllCookies retrieves all cookies for the current session
func (c *WebDriverClient) GetAllCookies(ctx context.Context) ([]map[string]interface{}, error) {
	if c.sessionID == "" {
		return nil, fmt.Errorf("no active session")
	}

	req, err := http.NewRequestWithContext(ctx, "GET",
		c.baseURL+"/session/"+c.sessionID+"/cookie", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get cookies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get cookies failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Value []map[string]interface{} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode cookies response: %w", err)
	}

	return result.Value, nil
}

// SetWindowSize sets the browser window size
func (c *WebDriverClient) SetWindowSize(ctx context.Context, width, height int) error {
	if c.sessionID == "" {
		return fmt.Errorf("no active session")
	}

	payload := map[string]interface{}{
		"width":  width,
		"height": height,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal window size payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/session/"+c.sessionID+"/window/rect", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create set window size request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set window size: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("set window size failed with status: %d", resp.StatusCode)
	}

	return nil
}

// CreateSession creates a new WebDriver session
func (c *WebDriverClient) CreateSession(ctx context.Context, capabilities map[string]interface{}) (*WebDriverSession, error) {
	payload := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"alwaysMatch": capabilities,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/session",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("session creation failed with status: %d", resp.StatusCode)
	}

	var sessionResp struct {
		Value WebDriverSession `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
		return nil, fmt.Errorf("failed to decode session response: %w", err)
	}

	c.sessionID = sessionResp.Value.SessionID
	return &sessionResp.Value, nil
}

// DeleteSession deletes the current WebDriver session
func (c *WebDriverClient) DeleteSession(ctx context.Context) error {
	if c.sessionID == "" {
		log.Println("WARN: attempted to delete session, but no active session exists")
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE",
		c.baseURL+"/session/"+c.sessionID, nil)
	if err != nil {
		log.Printf("WARN: failed to create delete request: %v\n", err)
		return nil
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("WARN: failed to delete session: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("WARN: session deletion failed with status: %d\n", resp.StatusCode)
		c.sessionID = ""
		return nil
	}

	c.sessionID = ""
	return nil
}

// NavigateOptions contains options for navigation
type NavigateOptions struct {
	WaitUntil string // "load" (default), "domcontentloaded", "networkidle"
}

// Navigate navigates to a URL with optional wait conditions
func (c *WebDriverClient) Navigate(ctx context.Context, url string, options *NavigateOptions) error {
	if c.sessionID == "" {
		return fmt.Errorf("no active session")
	}

	// Set defaults
	if options == nil {
		options = &NavigateOptions{
			WaitUntil: "load",
		}
	}
	if options.WaitUntil == "" {
		options.WaitUntil = "load"
	}

	payload := map[string]string{"url": url}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal navigate payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/session/"+c.sessionID+"/url", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create navigate request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("navigation failed with status: %d", resp.StatusCode)
	}

	// WebDriver's Navigate command waits for "load" by default
	// For other wait conditions, we need to poll
	switch options.WaitUntil {
	case "load":
		// Already waited by WebDriver
		return nil
	case "domcontentloaded":
		return c.waitForDOMContentLoaded(ctx)
	case "networkidle":
		return c.waitForNetworkIdle(ctx)
	default:
		return fmt.Errorf("invalid waitUntil option: %s", options.WaitUntil)
	}
}

// waitForDOMContentLoaded waits for the document to be interactive or complete
func (c *WebDriverClient) waitForDOMContentLoaded(ctx context.Context) error {
	script := `return document.readyState === 'interactive' || document.readyState === 'complete';`
	return c.pollForCondition(ctx, script)
}

// waitForNetworkIdle waits for network activity to settle
// This is a simplified implementation that waits for document.readyState === 'complete'
// and then waits an additional 500ms for any async operations
func (c *WebDriverClient) waitForNetworkIdle(ctx context.Context) error {
	// First wait for document to be complete
	script := `return document.readyState === 'complete';`
	err := c.pollForCondition(ctx, script)
	if err != nil {
		return err
	}

	// Then wait 500ms for any pending network activity to complete
	// This is a simple heuristic approach since WebDriver doesn't have native network idle detection
	time.Sleep(500 * time.Millisecond)

	return nil
}

// pollForCondition polls a JavaScript condition until it returns true or times out
func (c *WebDriverClient) pollForCondition(ctx context.Context, script string) error {
	interval := 100 * time.Millisecond
	timeout := 30 * time.Second // Fixed 30 second timeout
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		result, err := c.ExecuteScript(ctx, script, nil)
		if err != nil {
			return fmt.Errorf("failed to execute condition script: %w", err)
		}

		if boolResult, ok := result.(bool); ok && boolResult {
			return nil
		}

		time.Sleep(interval)
	}

	return fmt.Errorf("timeout waiting for condition after 30s")
}

// GetCurrentURL returns the current page URL
func (c *WebDriverClient) GetCurrentURL(ctx context.Context) (string, error) {
	if c.sessionID == "" {
		return "", fmt.Errorf("no active session")
	}

	req, err := http.NewRequestWithContext(ctx, "GET",
		c.baseURL+"/session/"+c.sessionID+"/url", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create get URL request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get current URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get URL failed with status: %d", resp.StatusCode)
	}

	var urlResp struct {
		Value string `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&urlResp); err != nil {
		return "", fmt.Errorf("failed to decode URL response: %w", err)
	}

	return urlResp.Value, nil
}

// GetTitle returns the current page title
func (c *WebDriverClient) GetTitle(ctx context.Context) (string, error) {
	if c.sessionID == "" {
		return "", fmt.Errorf("no active session")
	}

	req, err := http.NewRequestWithContext(ctx, "GET",
		c.baseURL+"/session/"+c.sessionID+"/title", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create get title request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get title: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get title failed with status: %d", resp.StatusCode)
	}

	var titleResp struct {
		Value string `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&titleResp); err != nil {
		return "", fmt.Errorf("failed to decode title response: %w", err)
	}

	return titleResp.Value, nil
}

// ExecuteScript executes JavaScript in the browser
func (c *WebDriverClient) ExecuteScript(ctx context.Context, script string, args []interface{}) (interface{}, error) {
	if c.sessionID == "" {
		return nil, fmt.Errorf("no active session")
	}

	// Ensure args is always an array, even if empty
	if args == nil {
		args = []interface{}{}
	}

	payload := map[string]interface{}{
		"script": script,
		"args":   args,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal script payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/session/"+c.sessionID+"/execute/sync", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create execute script request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute script: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Try to get error details from response body
		var errorBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorBody); err == nil {
			if value, ok := errorBody["value"].(map[string]interface{}); ok {
				if message, ok := value["message"].(string); ok {
					return nil, fmt.Errorf("script execution failed with status %d: %s", resp.StatusCode, message)
				}
			}
		}
		return nil, fmt.Errorf("script execution failed with status: %d", resp.StatusCode)
	}

	var scriptResp struct {
		Value interface{} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&scriptResp); err != nil {
		return nil, fmt.Errorf("failed to decode script response: %w", err)
	}

	return scriptResp.Value, nil
}

// FindElement finds an element using an auto-detected selector strategy
func (c *WebDriverClient) FindElement(ctx context.Context, selector string) (string, error) {
	// Use the new strategy-aware finder
	return c.FindElementWithStrategy(ctx, selector)
}

// FindElements returns the count of elements matching the selector
func (c *WebDriverClient) FindElements(ctx context.Context, selector string) (int, error) {
	elementIDs, err := c.FindAllElements(ctx, selector)
	if err != nil {
		return 0, err
	}
	return len(elementIDs), nil
}

// FindAllElements finds all elements matching the selector and returns their IDs
func (c *WebDriverClient) FindAllElements(ctx context.Context, selector string) ([]string, error) {
	parsed := ParseSelector(selector)

	if parsed.IsNative {
		return c.findAllElementsNative(ctx, string(parsed.Strategy), parsed.Value)
	}

	return c.findAllElementsCustom(ctx, parsed.Strategy, parsed.Value)
}

// findAllElementsNative uses WebDriver's native element finding for multiple elements
func (c *WebDriverClient) findAllElementsNative(ctx context.Context, strategy, value string) ([]string, error) {
	if c.sessionID == "" {
		return nil, fmt.Errorf("no active session")
	}

	payload := map[string]string{"using": strategy, "value": value}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal find elements payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/session/"+c.sessionID+"/elements", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create find elements request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to find elements: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("find elements failed with status: %d", resp.StatusCode)
	}

	var elementsResp struct {
		Value []struct {
			ElementID string `json:"element-6066-11e4-a52e-4f735466cecf"`
			ELEMENT   string `json:"ELEMENT"` // Fallback for older WebDriver
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&elementsResp); err != nil {
		return nil, fmt.Errorf("failed to decode elements response: %w", err)
	}

	elementIDs := make([]string, 0, len(elementsResp.Value))
	for _, elem := range elementsResp.Value {
		// Try W3C standard field first, fallback to older ELEMENT field
		if elem.ElementID != "" {
			elementIDs = append(elementIDs, elem.ElementID)
		} else if elem.ELEMENT != "" {
			elementIDs = append(elementIDs, elem.ELEMENT)
		}
	}

	return elementIDs, nil
}

// findAllElementsCustom uses JavaScript to find all elements with custom strategies
func (c *WebDriverClient) findAllElementsCustom(ctx context.Context, strategy SelectorStrategy, value string) ([]string, error) {
	script := generateAllSelectorScript(strategy, value)

	result, err := c.ExecuteScript(ctx, script, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to execute selector script: %w", err)
	}

	// Check if result is an array
	if result == nil {
		return []string{}, nil
	}

	// Handle array of element references
	if elemArray, ok := result.([]interface{}); ok {
		elementIDs := make([]string, 0, len(elemArray))
		for _, elem := range elemArray {
			if elemMap, ok := elem.(map[string]interface{}); ok {
				// Try W3C standard key
				if elemID, ok := elemMap["element-6066-11e4-a52e-4f735466cecf"].(string); ok {
					elementIDs = append(elementIDs, elemID)
				} else if elemID, ok := elemMap["ELEMENT"].(string); ok {
					elementIDs = append(elementIDs, elemID)
				}
			}
		}
		return elementIDs, nil
	}

	return []string{}, nil
}

// WaitForSelector waits for an element matching the selector to reach the specified state
func (c *WebDriverClient) WaitForSelector(ctx context.Context, selector, state string) error {
	if c.sessionID == "" {
		return fmt.Errorf("no active session")
	}

	// Generate the wait script based on state
	script := generateWaitScript(selector, state)

	// Create a context with fixed 30 second timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Poll until condition is met or timeout
	pollInterval := 100 * time.Millisecond
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctxWithTimeout.Done():
			return fmt.Errorf("timeout waiting for selector '%s' to be %s", selector, state)
		case <-ticker.C:
			// Execute the check script
			result, err := c.ExecuteScript(ctx, script, nil)
			if err != nil {
				// Continue polling on error
				continue
			}

			// Check if condition is met
			if satisfied, ok := result.(bool); ok && satisfied {
				return nil
			}
		}
	}
}

// generateWaitScript generates JavaScript to check element state
func generateWaitScript(selector, state string) string {
	parsed := ParseSelector(selector)

	// Build the element finding logic
	var findElementScript string
	if parsed.IsNative {
		switch parsed.Strategy {
		case StrategyCSSSelector:
			// Use querySelector for CSS selectors
			findElementScript = fmt.Sprintf(`document.querySelector("%s")`,
				strings.ReplaceAll(parsed.Value, `"`, `\"`))
		case StrategyXPath:
			// Use XPath evaluation for XPath selectors
			escapedXPath := strings.ReplaceAll(parsed.Value, `'`, `\'`)
			findElementScript = fmt.Sprintf(`document.evaluate('%s', document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue`, escapedXPath)
		default:
			// For other native strategies, use the selector script
			findElementScript = fmt.Sprintf(`(%s)`, generateSelectorScript(parsed.Strategy, parsed.Value))
		}
	} else {
		// Use custom selector script
		findElementScript = fmt.Sprintf(`(%s)`, generateSelectorScript(parsed.Strategy, parsed.Value))
	}

	// Build the state check based on the requested state
	switch state {
	case "attached":
		return fmt.Sprintf(`
			var element = %s;
			return element !== null && element !== undefined;
		`, findElementScript)

	case "detached":
		return fmt.Sprintf(`
			var element = %s;
			return element === null || element === undefined;
		`, findElementScript)

	case "visible":
		return fmt.Sprintf(`
			var element = %s;
			if (!element) return false;
			if (element.offsetWidth === 0 || element.offsetHeight === 0) return false;
			var style = window.getComputedStyle(element);
			return style.display !== 'none' && style.visibility !== 'hidden' && style.opacity !== '0';
		`, findElementScript)

	case "hidden":
		return fmt.Sprintf(`
			var element = %s;
			if (!element) return true;
			if (element.offsetWidth === 0 || element.offsetHeight === 0) return true;
			var style = window.getComputedStyle(element);
			return style.display === 'none' || style.visibility === 'hidden' || style.opacity === '0';
		`, findElementScript)

	default:
		// Default to visible
		return fmt.Sprintf(`
			var element = %s;
			if (!element) return false;
			if (element.offsetWidth === 0 || element.offsetHeight === 0) return false;
			var style = window.getComputedStyle(element);
			return style.display !== 'none' && style.visibility !== 'hidden';
		`, findElementScript)
	}
}

// ClickElement clicks an element by its ID
func (c *WebDriverClient) ClickElement(ctx context.Context, elementID string) error {
	if c.sessionID == "" {
		return fmt.Errorf("no active session")
	}

	elementRef := map[string]string{"element-6066-11e4-a52e-4f735466cecf": elementID}

	// Scroll, highlight, and click the element with detailed logging
	clickScript := `
		var element = arguments[0];
		if (!element) {
			return {success: false, error: "Element not found"};
		}

		// Get element info
		var info = {
			tagName: element.tagName,
			id: element.id,
			className: element.className,
			text: element.textContent ? element.textContent.substring(0, 50) : "",
			visible: element.offsetWidth > 0 && element.offsetHeight > 0,
			disabled: element.disabled,
			type: element.type
		};

		// Scroll into view
		element.scrollIntoView({behavior: 'instant', block: 'center', inline: 'center'});

		// Try to click
		try {
			element.click();
			return {success: true, info: info};
		} catch (e) {
			return {success: false, error: e.toString(), info: info};
		}
	`

	result, err := c.ExecuteScript(ctx, clickScript, []interface{}{elementRef})
	if err != nil {
		return fmt.Errorf("failed to click element via JavaScript: %w", err)
	}

	// Parse the result
	if resultMap, ok := result.(map[string]interface{}); ok {
		if success, ok := resultMap["success"].(bool); ok && !success {
			errorMsg := "unknown error"
			if errStr, ok := resultMap["error"].(string); ok {
				errorMsg = errStr
			}

			// Log element info if available
			if info, ok := resultMap["info"].(map[string]interface{}); ok {
				log.Printf("Click failed. Element info: %+v", info)
			}

			return fmt.Errorf("click failed: %s", errorMsg)
		}

		// Log success with element info
		if info, ok := resultMap["info"].(map[string]interface{}); ok {
			log.Printf("Click succeeded. Element info: %+v", info)
		}
	}

	return nil
}

// SendKeys sends text to an element
func (c *WebDriverClient) SendKeys(ctx context.Context, elementID, text string) error {
	if c.sessionID == "" {
		return fmt.Errorf("no active session")
	}

	payload := map[string]string{"text": text}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal send keys payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/session/"+c.sessionID+"/element/"+elementID+"/value", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create send keys request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send keys: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send keys failed with status: %d", resp.StatusCode)
	}

	return nil
}

// TakeScreenshot takes a screenshot of the current page, clipped to viewport size
func (c *WebDriverClient) TakeScreenshot(ctx context.Context) ([]byte, error) {
	if c.sessionID == "" {
		return nil, fmt.Errorf("no active session")
	}

	// Get viewport dimensions using JavaScript
	viewportScript := `
		return {
			width: window.innerWidth,
			height: window.innerHeight,
			devicePixelRatio: window.devicePixelRatio || 1
		};
	`

	viewportResult, err := c.ExecuteScript(ctx, viewportScript, []interface{}{})
	if err != nil {
		log.Printf("Warning: failed to get viewport dimensions: %v", err)
		// Fall back to full screenshot
		return c.takeFullScreenshot(ctx)
	}

	viewport, ok := viewportResult.(map[string]interface{})
	if !ok {
		log.Printf("Warning: unexpected viewport result type")
		return c.takeFullScreenshot(ctx)
	}

	// Extract dimensions
	var width, height int
	var dpr float64 = 1.0

	if w, ok := viewport["width"].(float64); ok {
		width = int(w)
	}
	if h, ok := viewport["height"].(float64); ok {
		height = int(h)
	}
	if d, ok := viewport["devicePixelRatio"].(float64); ok {
		dpr = d
	}

	log.Printf("Viewport: %dx%d, DPR: %.1f, Target size: %dx%d",
		width, height, dpr, int(float64(width)*dpr), int(float64(height)*dpr))

	// If we couldn't get dimensions, fall back to full screenshot
	if width == 0 || height == 0 {
		log.Printf("Warning: invalid viewport dimensions, using full screenshot")
		return c.takeFullScreenshot(ctx)
	}

	// Take full screenshot first
	fullScreenshot, err := c.takeFullScreenshot(ctx)
	if err != nil {
		return nil, err
	}

	// Crop to viewport size accounting for device pixel ratio
	targetWidth := int(float64(width) * dpr)
	targetHeight := int(float64(height) * dpr)

	croppedScreenshot, err := c.cropImage(fullScreenshot, targetWidth, targetHeight)
	if err != nil {
		log.Printf("Warning: failed to crop screenshot: %v, returning full screenshot", err)
		return fullScreenshot, nil
	}

	log.Printf("Successfully cropped screenshot to %dx%d", targetWidth, targetHeight)
	return croppedScreenshot, nil
}

// takeFullScreenshot takes a full page screenshot
func (c *WebDriverClient) takeFullScreenshot(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		c.baseURL+"/session/"+c.sessionID+"/screenshot", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create screenshot request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to take screenshot: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("screenshot failed with status: %d", resp.StatusCode)
	}

	var screenshotResp struct {
		Value string `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&screenshotResp); err != nil {
		return nil, fmt.Errorf("failed to decode screenshot response: %w", err)
	}

	// Decode base64 screenshot data
	decoded, err := base64.StdEncoding.DecodeString(screenshotResp.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 screenshot: %w", err)
	}

	return decoded, nil
}

// cropImage crops a PNG image to the specified width and height
func (c *WebDriverClient) cropImage(imageData []byte, width, height int) ([]byte, error) {
	img, err := decodePNG(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %w", err)
	}

	// Get actual image dimensions
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	log.Printf("Original image: %dx%d, Requested crop: %dx%d", imgWidth, imgHeight, width, height)

	// If requested size is larger or equal to actual size, return original
	if width >= imgWidth && height >= imgHeight {
		log.Printf("Requested size >= actual size, returning original")
		return imageData, nil
	}

	// Crop to requested size (top-left corner)
	croppedWidth := width
	croppedHeight := height
	if croppedWidth > imgWidth {
		croppedWidth = imgWidth
	}
	if croppedHeight > imgHeight {
		croppedHeight = imgHeight
	}

	log.Printf("Cropping to: %dx%d", croppedWidth, croppedHeight)

	// Create cropped image
	croppedImg := cropImageRect(img, 0, 0, croppedWidth, croppedHeight)

	// Encode back to PNG
	return encodePNG(croppedImg)
}

// Helper functions for image manipulation
func decodePNG(data []byte) (*image.RGBA, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	// Convert to RGBA
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	return rgba, nil
}

func encodePNG(img *image.RGBA) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func cropImageRect(img *image.RGBA, x, y, width, height int) *image.RGBA {
	cropped := image.NewRGBA(image.Rect(0, 0, width, height))

	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			cropped.Set(dx, dy, img.At(x+dx, y+dy))
		}
	}

	return cropped
}
