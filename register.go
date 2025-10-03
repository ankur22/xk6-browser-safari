package browser_safari

import "go.k6.io/k6/js/modules"

const importPath = "k6/x/browser_safari"

func init() {
	modules.Register(importPath, new(rootModule))
}
