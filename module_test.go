package browser_safari

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
	"go.k6.io/k6/js/modulestest"
)

func Test_module(t *testing.T) { //nolint:tparallel
	t.Parallel()

	runtime := modulestest.NewRuntime(t)

	err := runtime.SetupModuleSystem(map[string]any{importPath: new(rootModule)}, nil, nil)
	require.NoError(t, err)

	_, err = runtime.RunOnEventLoop(`let mod = require("` + importPath + `")`)
	require.NoError(t, err)

	tests := []struct {
		name  string
		check string
	}{
		{name: "browser", check: `typeof mod.browser === "object"`},
		{name: "browser.newPage()", check: `typeof mod.browser.newPage === "function"`},
		{name: "browser.close()", check: `typeof mod.browser.close === "function"`},
	}
	for _, tt := range tests { //nolint:paralleltest
		t.Run(tt.name, func(t *testing.T) {
			got, err := runtime.RunOnEventLoop(tt.check)

			require.NoError(t, err)
			require.True(t, got.ToBoolean())
		})
	}
}
