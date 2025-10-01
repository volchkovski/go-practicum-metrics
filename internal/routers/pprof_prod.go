//go:build !debug

package routers

import "github.com/go-chi/chi/v5"

// addDebugRoutes is a no-op when not building with debug tag.
// To enable profiling endpoints, build with: go build -tags debug
func addDebugRoutes(r chi.Router) {
	// No debug routes added in production builds
}
