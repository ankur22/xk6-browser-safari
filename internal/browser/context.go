package browser

import (
	"context"
	"fmt"

	"github.com/grafana/sobek"
	"go.k6.io/k6/js/modules"
)

// BrowserContext represents a browser context (conceptual layer over WebDriver sessions)
type BrowserContext struct {
	browser *Browser
	vu      modules.VU
	options map[string]interface{} // Store context options (e.g., viewport)
	pages   []*Page                // Track pages created in this context
}

// NewPage creates a new page in this browser context
func (bc *BrowserContext) NewPage() (*sobek.Promise, error) {
	// Delegate to browser's NewPage implementation with stored options
	if bc.options != nil {
		return bc.browser.NewPage(bc.options)
	}
	return bc.browser.NewPage()
}

// Cookies returns all cookies for the current context
func (bc *BrowserContext) Cookies() (*sobek.Promise, error) {
	return Promise(bc.vu, func() (interface{}, error) {
		ctx := context.Background()

		// Get cookies from the WebDriver session
		// If there's no active session, this will return an error
		cookies, err := bc.browser.Client.GetAllCookies(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get cookies: %w", err)
		}

		return cookies, nil
	}), nil
}
