package browser

import (
	"context"
	"fmt"

	"github.com/grafana/sobek"
	"go.k6.io/k6/js/modules"
)

// Locator represents a way to find element(s) on the page at any moment
type Locator struct {
	page      *Page
	selector  string
	elementID string // If set, this locator refers to a specific element
	vu        modules.VU
}

// Click clicks on the element matched by the locator
func (l *Locator) Click() (*sobek.Promise, error) {
	return Promise(l.vu, func() (interface{}, error) {
		if l.page.client == nil {
			return nil, fmt.Errorf("browser session not initialized")
		}

		ctx := context.Background()

		// If we already have a specific element ID, use it
		var elementID string
		var err error
		if l.elementID != "" {
			elementID = l.elementID
		} else {
			// Otherwise, find the element now
			elementID, err = l.page.client.FindElement(ctx, l.selector)
			if err != nil {
				return nil, fmt.Errorf("failed to find element with selector '%s': %w", l.selector, err)
			}
		}

		err = l.page.client.ClickElement(ctx, elementID)
		if err != nil {
			return nil, fmt.Errorf("failed to click element: %w", err)
		}

		return nil, nil
	}), nil
}

// Count returns the number of elements matching the locator
func (l *Locator) Count() (*sobek.Promise, error) {
	return Promise(l.vu, func() (interface{}, error) {
		if l.page.client == nil {
			return nil, fmt.Errorf("browser session not initialized")
		}

		ctx := context.Background()
		count, err := l.page.client.FindElements(ctx, l.selector)
		if err != nil {
			return nil, fmt.Errorf("failed to find elements with selector '%s': %w", l.selector, err)
		}

		return count, nil
	}), nil
}

// All returns all elements matching the locator as an array of Locators
func (l *Locator) All() (*sobek.Promise, error) {
	return Promise(l.vu, func() (interface{}, error) {
		if l.page.client == nil {
			return nil, fmt.Errorf("browser session not initialized")
		}

		ctx := context.Background()
		elementIDs, err := l.page.client.FindAllElements(ctx, l.selector)
		if err != nil {
			return nil, fmt.Errorf("failed to find elements with selector '%s': %w", l.selector, err)
		}

		// Create a locator for each specific element
		locators := make([]*Locator, len(elementIDs))
		for i, elementID := range elementIDs {
			locators[i] = &Locator{
				page:      l.page,
				selector:  l.selector,
				elementID: elementID,
				vu:        l.vu,
			}
		}

		return locators, nil
	}), nil
}

// WaitFor waits for the locator to satisfy the given state
func (l *Locator) WaitFor(options map[string]interface{}) (*sobek.Promise, error) {
	return Promise(l.vu, func() (interface{}, error) {
		if l.page.client == nil {
			return nil, fmt.Errorf("browser session not initialized")
		}

		// Parse state option (default: "visible")
		state := "visible"
		if options != nil {
			if stateVal, ok := options["state"].(string); ok {
				state = stateVal
			}
		}

		ctx := context.Background()
		err := l.page.client.WaitForSelector(ctx, l.selector, state)
		if err != nil {
			return nil, fmt.Errorf("waitFor failed for selector '%s': %w", l.selector, err)
		}

		return nil, nil
	}), nil
}

// TextContent returns the text content of the element
func (l *Locator) TextContent() (*sobek.Promise, error) {
	return Promise(l.vu, func() (interface{}, error) {
		if l.page.client == nil {
			return nil, fmt.Errorf("browser session not initialized")
		}

		ctx := context.Background()

		// If we already have a specific element ID, use it
		var elementID string
		var err error
		if l.elementID != "" {
			elementID = l.elementID
		} else {
			// Otherwise, find the element now
			elementID, err = l.page.client.FindElement(ctx, l.selector)
			if err != nil {
				return nil, fmt.Errorf("failed to find element with selector '%s': %w", l.selector, err)
			}
		}

		// Get the text content using JavaScript
		script := `
			var element = arguments[0];
			if (!element) return null;
			return element.textContent;
		`

		elementRef := map[string]string{"element-6066-11e4-a52e-4f735466cecf": elementID}
		result, err := l.page.client.ExecuteScript(ctx, script, []interface{}{elementRef})
		if err != nil {
			return nil, fmt.Errorf("failed to get text content: %w", err)
		}

		return result, nil
	}), nil
}

// Type types text into the element character by character
func (l *Locator) Type(text string, options ...map[string]interface{}) (*sobek.Promise, error) {
	return Promise(l.vu, func() (interface{}, error) {
		if l.page.client == nil {
			return nil, fmt.Errorf("browser session not initialized")
		}

		ctx := context.Background()

		// If we already have a specific element ID, use it
		var elementID string
		var err error
		if l.elementID != "" {
			elementID = l.elementID
		} else {
			// Otherwise, find the element now
			elementID, err = l.page.client.FindElement(ctx, l.selector)
			if err != nil {
				return nil, fmt.Errorf("failed to find element with selector '%s': %w", l.selector, err)
			}
		}

		// Parse delay option (default: 0ms between keystrokes)
		// Note: delay is acknowledged but not implemented due to WebDriver limitations
		_ = 0 // delay placeholder
		if len(options) > 0 && options[0] != nil {
			if delayVal, ok := options[0]["delay"].(float64); ok {
				_ = int(delayVal)
			} else if delayVal, ok := options[0]["delay"].(int); ok {
				_ = delayVal
			}
		}

		// Use WebDriver's SendKeys command to type text
		err = l.page.client.SendKeys(ctx, elementID, text)
		if err != nil {
			return nil, fmt.Errorf("failed to type text: %w", err)
		}

		// Note: WebDriver's SendKeys sends all text at once
		// Per-character delays are not supported natively by WebDriver

		return nil, nil
	}), nil
}
