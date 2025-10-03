package browser

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/grafana/sobek"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/js/promises"
)

//go:embed injection_script.js
var injectionScript string

var (
	safariDriverCmd  *exec.Cmd
	safariDriverMu   sync.Mutex
	safariDriverRefs int
)

// StartSafariDriver starts safaridriver if it's not already running
func StartSafariDriver() error {
	safariDriverMu.Lock()
	defer safariDriverMu.Unlock()

	// If already running, just increment reference count
	if safariDriverCmd != nil && safariDriverCmd.Process != nil {
		safariDriverRefs++
		return nil
	}

	// Check if port 4444 is already in use
	if isPortInUse(4444) {
		// Assume safaridriver is already running externally
		safariDriverRefs++
		return nil
	}

	// Start safaridriver
	cmd := exec.Command("safaridriver", "--port", "4444")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start safaridriver: %w", err)
	}

	safariDriverCmd = cmd
	safariDriverRefs = 1

	// Wait for safaridriver to be ready
	if err := waitForPort(4444, 10*time.Second); err != nil {
		cmd.Process.Kill()
		safariDriverCmd = nil
		return fmt.Errorf("safaridriver did not become ready: %w", err)
	}

	return nil
}

// stopSafariDriver decrements the reference count and stops safaridriver if no more references
func stopSafariDriver() {
	safariDriverMu.Lock()
	defer safariDriverMu.Unlock()

	if safariDriverRefs > 0 {
		safariDriverRefs--
	}

	// Only stop if we started it and there are no more references
	if safariDriverRefs == 0 && safariDriverCmd != nil && safariDriverCmd.Process != nil {
		safariDriverCmd.Process.Kill()
		safariDriverCmd.Wait()
		safariDriverCmd = nil
	}
}

// isPortInUse checks if a TCP port is in use
func isPortInUse(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 100*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// waitForPort waits for a TCP port to become available
func waitForPort(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if isPortInUse(port) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("port %d did not become available within %v", port, timeout)
}

// Viewport represents the browser viewport dimensions
type Viewport struct {
	Width  int
	Height int
}

// Browser represents a Safari browser instance
type Browser struct {
	VU     modules.VU
	Client *WebDriverClient
}

// NewContext creates a new browser context with optional configuration
func (b *Browser) NewContext(options ...map[string]interface{}) *BrowserContext {
	var opts map[string]interface{}
	if len(options) > 0 {
		opts = options[0]
	}

	return &BrowserContext{
		browser: b,
		vu:      b.VU,
		options: opts,
	}
}

// NewPage creates a new page in the browser
func (b *Browser) NewPage(options ...map[string]interface{}) (*sobek.Promise, error) {
	return Promise(b.VU, func() (any, error) {
		ctx := context.Background()

		// Parse viewport options
		viewport := &Viewport{Width: 1280, Height: 720} // Default viewport
		if len(options) > 0 && options[0] != nil {
			if viewportOpt, ok := options[0]["viewport"].(map[string]interface{}); ok {
				if width, ok := viewportOpt["width"].(float64); ok {
					viewport.Width = int(width)
				}
				if height, ok := viewportOpt["height"].(float64); ok {
					viewport.Height = int(height)
				}
			}
		}

		// Create a new WebDriver session with viewport
		capabilities := map[string]interface{}{
			"browserName":             "Safari",
			"safari:devicePixelRatio": 1.0, // Force DPR to 1 for consistent screenshots
		}

		session, err := b.Client.CreateSession(ctx, capabilities)
		if err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}

		page := &Page{
			vu:      b.VU,
			client:  b.Client,
			session: session,
		}

		// Set the window size to match viewport
		// Add extra height to account for Safari's browser chrome (address bar, tabs, etc.)
		// Safari's chrome is typically around 52-60 pixels
		windowHeight := viewport.Height + 52
		if err := b.Client.SetWindowSize(ctx, viewport.Width, windowHeight); err != nil {
			fmt.Printf("WARN: failed to set window size: %v\n", err)
		}

		// Inject the initialization script
		if err := page.injectScript(ctx); err != nil {
			// Log warning but don't fail page creation
			fmt.Printf("WARN: failed to inject initialization script: %v\n", err)
		}

		return page, nil
	}), nil
}

// Close closes the browser and all its pages
func (b *Browser) Close() (*sobek.Promise, error) {
	return Promise(b.VU, func() (any, error) {
		ctx := context.Background()
		err := b.Client.DeleteSession(ctx)

		// Decrement safaridriver reference count
		stopSafariDriver()

		return nil, err
	}), nil
}

// Page represents a browser page
type Page struct {
	vu      modules.VU
	client  *WebDriverClient
	session *WebDriverSession
}

// injectScript injects the initialization script into the page
func (p *Page) injectScript(ctx context.Context) error {
	if p.client == nil {
		return fmt.Errorf("browser session not initialized")
	}

	// Execute the embedded injection script
	_, err := p.client.ExecuteScript(ctx, injectionScript, nil)
	return err
}

// Goto navigates to a URL with optional wait conditions
func (p *Page) Goto(url string, options map[string]interface{}) (*sobek.Promise, error) {
	if p.client == nil {
		return nil, fmt.Errorf("browser session not initialized")
	}

	return Promise(p.vu, func() (any, error) {
		ctx := context.Background()

		// Parse options
		var navOptions *NavigateOptions
		if options != nil {
			navOptions = &NavigateOptions{
				WaitUntil: "load",
			}

			if waitUntil, ok := options["waitUntil"].(string); ok {
				navOptions.WaitUntil = waitUntil
			}
		}

		err := p.client.Navigate(ctx, url, navOptions)
		if err != nil {
			return nil, err
		}

		// Re-inject the script after navigation
		if err := p.injectScript(ctx); err != nil {
			// Log warning but don't fail navigation
			fmt.Printf("WARN: failed to inject script after navigation: %v\n", err)
		}

		return nil, nil
	}), nil
}

// URL returns the current page URL
func (p *Page) URL() string {
	if p.client == nil {
		return ""
	}

	ctx := context.Background()
	url, err := p.client.GetCurrentURL(ctx)
	if err != nil {
		return ""
	}
	return url
}

// Locator creates a locator for the given selector (synchronous method)
func (p *Page) Locator(selector string) *Locator {
	return &Locator{
		page:     p,
		selector: selector,
		vu:       p.vu,
	}
}

// Title returns the current page title
func (p *Page) Title() (*sobek.Promise, error) {
	if p.client == nil {
		return nil, fmt.Errorf("browser session not initialized")
	}

	return Promise(p.vu, func() (any, error) {
		ctx := context.Background()
		title, err := p.client.GetTitle(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get title: %w", err)
		}
		return title, nil
	}), nil
}

// Evaluate executes JavaScript and returns the result
func (p *Page) Evaluate(script string) (*sobek.Promise, error) {
	if p.client == nil {
		return nil, fmt.Errorf("browser session not initialized")
	}

	return Promise(p.vu, func() (any, error) {
		ctx := context.Background()
		result, err := p.client.ExecuteScript(ctx, script, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to execute script: %w", err)
		}
		return result, nil
	}), nil
}

// Click clicks an element by CSS selector
func (p *Page) Click(selector string) (*sobek.Promise, error) {
	if p.client == nil {
		return nil, fmt.Errorf("browser session not initialized")
	}

	return Promise(p.vu, func() (any, error) {
		ctx := context.Background()
		elementID, err := p.client.FindElement(ctx, selector)
		if err != nil {
			return nil, fmt.Errorf("failed to find element: %w", err)
		}

		err = p.client.ClickElement(ctx, elementID)
		if err != nil {
			return nil, fmt.Errorf("failed to click element: %w", err)
		}

		return nil, nil
	}), nil
}

// Fill fills an input field with text
func (p *Page) Fill(selector, text string) (*sobek.Promise, error) {
	if p.client == nil {
		return nil, fmt.Errorf("browser session not initialized")
	}

	return Promise(p.vu, func() (any, error) {
		ctx := context.Background()
		elementID, err := p.client.FindElement(ctx, selector)
		if err != nil {
			return nil, fmt.Errorf("failed to find element: %w", err)
		}

		err = p.client.SendKeys(ctx, elementID, text)
		if err != nil {
			return nil, fmt.Errorf("failed to send keys: %w", err)
		}

		return nil, nil
	}), nil
}

// Screenshot takes a screenshot of the current page
func (p *Page) Screenshot(options map[string]interface{}) (*sobek.Promise, error) {
	if p.client == nil {
		return nil, fmt.Errorf("browser session not initialized")
	}

	return Promise(p.vu, func() (any, error) {
		ctx := context.Background()
		screenshotData, err := p.client.TakeScreenshot(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to take screenshot: %w", err)
		}

		// If path is provided, write the screenshot to file
		if pathValue, exists := options["path"]; exists {
			if pathStr, ok := pathValue.(string); ok {
				err = os.WriteFile(pathStr, screenshotData, 0644)
				if err != nil {
					return nil, fmt.Errorf("failed to write screenshot to file: %w", err)
				}
			}
		}

		// Always return the buffer, like Playwright does
		return screenshotData, nil
	}), nil
}

// WaitForTimeout waits for the specified number of milliseconds
func (p *Page) WaitForTimeout(milliseconds int) (*sobek.Promise, error) {
	return Promise(p.vu, func() (interface{}, error) {
		duration := time.Duration(milliseconds) * time.Millisecond
		time.Sleep(duration)
		return nil, nil
	}), nil
}

// Close closes the page
func (p *Page) Close() (*sobek.Promise, error) {
	if p.client == nil {
		return nil, fmt.Errorf("browser session not initialized")
	}

	return Promise(p.vu, func() (any, error) {
		ctx := context.Background()
		err := p.client.DeleteSession(ctx)

		// Decrement safaridriver reference count
		stopSafariDriver()

		return nil, err
	}), nil
}

// PromisifiedFunc is a type of the function to run as a promise.
type PromisifiedFunc func() (result any, reason error)

// Promise runs fn in a goroutine and returns a new sobek.Promise.
//   - If fn returns a nil error, resolves the promise with the
//     first result value fn returns.
//   - Otherwise, rejects the promise with the error fn returns.
func Promise(vu modules.VU, fn PromisifiedFunc) *sobek.Promise {
	p, resolve, reject := promises.New(vu)
	go func() {
		v, err := fn()
		if err != nil {
			reject(err)
			return
		}
		resolve(v)
	}()

	return p
}

type ctxKey int

const (
	ctxKeyVU ctxKey = iota
	ctxKeyPid
	ctxKeyCustomK6Metrics
)

// GetVU returns the attached k6 VU instance from ctx, which can be used to
// retrieve the sobek runtime and other k6 objects relevant to the currently
// executing VU.
// See https://github.com/grafana/k6/blob/v0.37.0/js/initcontext.go#L168-L186
func GetVU(ctx context.Context) modules.VU {
	v := ctx.Value(ctxKeyVU)
	if vu, ok := v.(modules.VU); ok {
		return vu
	}
	return nil
}
