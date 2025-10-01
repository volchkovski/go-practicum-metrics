//go:build debug

package routers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// addDebugRoutes adds profiling endpoints to the router.
// This function is only available when building with the debug tag.
// Use: go build -tags debug to enable profiling endpoints.
func addDebugRoutes(r chi.Router) {
	r.Mount("/debug", middleware.Profiler())
}
