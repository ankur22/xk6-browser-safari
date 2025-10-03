package browser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.k6.io/k6/js/modulestest"
)

func TestBrowserContext(t *testing.T) {
	t.Parallel()

	runtime := modulestest.NewRuntime(t)

	browser := &Browser{
		VU:     runtime.VU,
		Client: NewWebDriverClient("http://localhost:4444"),
	}

	// Test without options
	context := browser.NewContext()
	require.NotNil(t, context)
	require.Equal(t, browser, context.browser)
	require.Equal(t, runtime.VU, context.vu)
	require.Nil(t, context.options)

	// Test with options
	options := map[string]interface{}{
		"viewport": map[string]interface{}{
			"width":  1920,
			"height": 1080,
		},
	}
	contextWithOptions := browser.NewContext(options)
	require.NotNil(t, contextWithOptions)
	require.NotNil(t, contextWithOptions.options)
	require.Equal(t, options, contextWithOptions.options)
}

func TestBrowserContextNewPage(t *testing.T) {
	t.Parallel()

	runtime := modulestest.NewRuntime(t)

	browser := &Browser{
		VU:     runtime.VU,
		Client: NewWebDriverClient("http://localhost:4444"),
	}

	context := browser.NewContext()

	// NewPage should return a promise
	promise, err := context.NewPage()
	require.NoError(t, err)
	require.NotNil(t, promise)
}

func TestBrowserContextCookies(t *testing.T) {
	t.Parallel()

	runtime := modulestest.NewRuntime(t)

	browser := &Browser{
		VU:     runtime.VU,
		Client: NewWebDriverClient("http://localhost:4444"),
	}

	context := browser.NewContext()

	// Cookies should return a promise
	promise, err := context.Cookies()
	require.NoError(t, err)
	require.NotNil(t, promise)
}
