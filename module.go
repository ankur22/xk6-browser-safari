// Package browser_safari contains the xk6-browser-safari extension.
package browser_safari

import (
	"xk6-browser-safari/internal/browser"

	"go.k6.io/k6/js/modules"
)

type rootModule struct{}

func (*rootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &module{vu}
}

type module struct {
	vu modules.VU
}

func (m *module) Exports() modules.Exports {
	// Start safaridriver when module loads
	if err := browser.StartSafariDriver(); err != nil {
		// Log error but don't fail module loading
		// The error will surface when trying to create a page
	}

	// Create and return the browser instance directly
	b := &browser.Browser{
		VU:     m.vu,
		Client: browser.NewWebDriverClient("http://localhost:4444"),
	}

	return modules.Exports{
		Named: map[string]any{
			"browser":            b,
			"compareScreenshots": browser.CompareImages,
			"createDiffImage":    browser.CreateDiffImage,
		},
	}
}

var _ modules.Module = (*rootModule)(nil)
