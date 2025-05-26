package runtime

import (
	"context"
	"encoding/json"
	"errors" // Added for ErrResourceNotFound
	"fmt"    // Added for formatting errors in adapter
	"io"

	"github.com/hyperleex/zenmcp/registry" // Assuming registry is needed for NewRouter
	// We need to refer to Router from router.go.
	// To avoid import cycle if router.go needs to import Server from here,
	// we might need to rethink. For now, let's assume this direct import is fine.
	// Or, the router.Router type might need to be passed in or defined in a common package.
	// For this step, we'll assume runtime.Router is accessible.
)

// ErrResourceNotFound is returned when a resource URI is not found.
var ErrResourceNotFound = errors.New("resource not found")

// ResourcesReadParams defines the parameters for the "resources/read" method.
type ResourcesReadParams struct {
	URI string `json:"uri"`
}

// Resource defines a discoverable resource.
type Resource struct {
	URI      string                        `json:"uri"`
	Name     string                        `json:"name,omitempty"`
	MimeType string                        `json:"mimeType,omitempty"`
	Reader   func() (io.ReadCloser, error) `json:"-"` // Exclude reader from JSON responses
}

// Server is the main mcpgo server.
type Server struct {
	resourceProviders map[string]func(context.Context) ([]Resource, error)
	router            *Router // Changed from placeholder comment
	// other server fields
}

// NewServer creates a new server instance.
// It now requires a registry to initialize the router.
func NewServer(reg *registry.Registry) *Server {
	s := &Server{
		resourceProviders: make(map[string]func(context.Context) ([]Resource, error)),
		router:            NewRouter(reg), // Initialize the router
	}

	// Register the resources/list handler
	// The handler from router.go expects: func(ctx *Context, params json.RawMessage) (interface{}, error)
	// Our s.handleResourcesList is: func(ctx context.Context) (any, error)
	// We need an adapter. Assuming router.Context is compatible with context.Context or we define it.
	// For now, let's assume router.Context is what the router expects.
	// If router.Context is just an alias for context.Context, this is simpler.
	// Let's assume for now that router.Context is a distinct type and we need to adapt.
	// However, looking at router.go, RequestHandler uses `*Context` which is not defined in the snippet.
	// Let's assume `runtime.Context` is the type defined and used in router.go.
	// If `runtime.Context` has a `context.Context` embedded, we can use that.
	// For now, let's write an adapter assuming `ctx *Context` in `RequestHandler` can provide `context.Context`.

	// The following large block assigning an inline function to s.router.handlers["resources/list"]
	// is removed. It contained an incorrect call to s.handleResourcesList(ctx, params)
	// which caused a linter error (runtime/resource.go:115:37).
	// This entire inline handler definition was superseded by the assignment
	// to `resourceListHandlerAdapter` which correctly handles the call.
	/*
	s.router.handlers["resources/list"] = func(ctx *Context, params json.RawMessage) (interface{}, error) {
		// Assuming ctx (*runtime.Context) can provide a standard context.Context if needed
		// or handleResourcesList needs to be adapted to take *runtime.Context.
		// For now, we'll pass ctx directly if its underlying type is context.Context,
		// otherwise this needs adjustment based on actual *Context definition.
		// Given handleResourcesList takes context.Context, and router.RequestHandler takes *runtime.Context,
		// this will only work if runtime.Context is context.Context or wraps it and can be easily extracted.
		// Let's assume for the purpose of this step that `ctx.StdContext()` exists or similar,
		// or that handleResourcesList is changed.
		// Simpler: adapt s.handleResourcesList to match signature if *Context is not just context.Context.
		// For now, let's assume we can call it directly if Context is an alias or wrapper.
		// The problem states "You might need to adjust function signatures".
		// Let's make handleResourcesList match the expected signature more closely first.
		// No, let's make an adapter as originally planned.

		// The router expects `(interface{}, error)` and `handleResourcesList` returns `(any, error)`.
		// `any` is an alias for `interface{}`, so this is compatible.
		// The main issue is the context type and unused params.
		// `handleResourcesList` expects `context.Context`.
		// The router provides `*runtime.Context`. We need to resolve this.

		// Simplest path: modify handleResourcesList to accept (*runtime.Context, json.RawMessage)
		// Or, assume runtime.Context is a struct that embeds context.Context, e.g., type Context struct { context.Context }
		// Let's try to keep handleResourcesList's core logic clean and use an adapter.

		// Adapter:
		// We need to know the definition of runtime.Context.
		// If runtime.Context is just `type Context context.Context`, then ctx itself is context.Context.
		// If not, we need a way to get context.Context from runtime.Context.
		// Let's assume `ctx.Std()` gives `context.Context`. This is a guess.
		// If `runtime.Context` is not defined, this will be a compile error.
		// The original `router.go` uses `ctx *Context`.
		// Let's make the adapter and assume `ctx` (type `*Context` from router.go) can be passed to `s.handleResourcesList`.
		// This implies `*Context` must be `context.Context` or satisfy it.
		// This is unlikely.

		// More robust adapter:
		// The handler in router.go is: func(ctx *Context, params json.RawMessage) (interface{}, error)
		// Our handler is: func(s *Server) handleResourcesList(stdCtx context.Context) (any, error)
		// We need to call s.handleResourcesList.
		// The router's RequestHandler is NOT a method of Server. It's a standalone function.
		// So, the registration should be:
		// s.router.handlers["resources/list"] = s.adaptedResourceListHandler()
		// OR: s.router.handlers["resources/list"] = func(routerCtx *Context, params json.RawMessage) (interface{}, error) {
		//    return s.handleResourcesList(context.Background()) // Or derive context from routerCtx if possible
		// }
		// This assumes s is available in this scope (it is, via closure).

		// Let's assume `router.Context` is the one defined in `router.go` (though its definition is missing).
		// And `resource.go`'s `handleResourcesList` takes `context.Context`.
		// We need to bridge this.
		// Simplest assumption for now: `router.Context` is, or can provide, `context.Context`.
		// If `router.Context` is not `context.Context`, this adapter will need modification.
		// Let's assume it's `context.Context` for now, as `Context` is a common name for it.
		return s.handleResourcesList(ctx, params) // This won't work because of signature mismatch.

		// Corrected adapter:
		// The router will call this function with *runtime.Context and json.RawMessage.
		// s.handleResourcesList expects context.Context and no json.RawMessage.
		// We need to ensure that *runtime.Context can be converted to context.Context.
		// For now, let's assume `ctx` (*runtime.Context) can be used as `context.Context`.
		// This is a strong assumption. The alternative is to change handleResourcesList signature.
		// Let's try to change handleResourcesList to match what the router provides.
	}
	// The above registration is problematic because of signature mismatch.
	// Let's redefine handleResourcesList slightly or use a proper adapter.
	*/

	// Option A: Change handleResourcesList signature (less ideal if it's a public API of Server)
	// Option B: Adapter (preferred)

	// Adapter function:
	// router.go defines: type RequestHandler func(ctx *Context, params json.RawMessage) (interface{}, error)
	// And also: type Context = context.Context
	// So, the `ctx` parameter in RequestHandler is of type `*context.Context`.
	// Our `s.handleResourcesList` expects `context.Context`.
	// The adapter needs to correctly dereference the pointer.
	resourceListHandlerAdapter := func(routerCtxPointer *Context, params json.RawMessage) (interface{}, error) {
		// routerCtxPointer is of type *runtime.Context, which should be *context.Context due to the alias in router.go.
		var stdCtx context.Context
		if routerCtxPointer != nil {
			stdCtx = *routerCtxPointer // Dereference the pointer to get the context.Context value
		} else {
			// This case should ideally not happen if the router always provides a valid context pointer.
			// Using context.Background() as a fallback.
			// Consider logging an error here or returning one if a nil context pointer is unexpected.
			stdCtx = context.Background()
		}
		// params json.RawMessage is ignored by s.handleResourcesList, which is fine.
		return s.handleResourcesList(stdCtx)
	}
	s.router.handlers["resources/list"] = resourceListHandlerAdapter
	// End of registration for resources/list

	// Register the resources/read handler
	s.router.handlers["resources/read"] = s.resourceReadHandlerAdapter
	// End of registration for resources/read

	return s
}

// resourceReadHandlerAdapter adapts the handleResourcesRead method to the router's RequestHandler signature.
func (s *Server) resourceReadHandlerAdapter(routerCtxPointer *Context, rawParams json.RawMessage) (interface{}, error) {
	var params ResourcesReadParams
	if err := json.Unmarshal(rawParams, &params); err != nil {
		// TODO: Return a proper JSON-RPC error type/code
		return nil, fmt.Errorf("invalid params for resources/read: %w", err)
	}

	var stdCtx context.Context
	if routerCtxPointer != nil {
		stdCtx = *routerCtxPointer // Dereference *context.Context to context.Context
	} else {
		// This case should ideally not happen if the router always provides a valid context pointer.
		stdCtx = context.Background()
	}

	return s.handleResourcesRead(stdCtx, params)
}

// handleResourcesRead handles the "resources/read" JSON-RPC method.
// It finds a resource by URI and returns its content.
func (s *Server) handleResourcesRead(ctx context.Context, params ResourcesReadParams) (interface{}, error) {
	for _, provider := range s.resourceProviders {
		resources, err := provider(ctx)
		if err != nil {
			// TODO: Log error from provider? Or collect and return multiple errors?
			// For now, skip provider if it errors, or potentially return the error.
			// Depending on desired behavior, could accumulate errors or return immediately.
			// Let's log and continue for now, to give other providers a chance.
			// log.Printf("error from resource provider: %v", err) // Placeholder for actual logging
			continue
		}

		for _, resource := range resources {
			if resource.URI == params.URI {
				if resource.Reader == nil {
					return nil, fmt.Errorf("resource %s has no reader defined", params.URI)
				}
				reader, err := resource.Reader()
				if err != nil {
					return nil, fmt.Errorf("failed to create reader for resource %q: %w", params.URI, err)
				}
				// ---- START FIX ----
				if reader == nil {
					// This case should ideally not happen if providers are well-behaved,
					// but as a safeguard, treat as resource not readable or content not available.
					return nil, fmt.Errorf("reader for resource %q is nil, but no error was reported by the reader function", params.URI)
				}
				// ---- END FIX ----
				defer reader.Close()

				data, err := io.ReadAll(reader)
				if err != nil {
					return nil, fmt.Errorf("failed to read data from resource %s: %w", params.URI, err)
				}
				// Data could be returned as string(data) if text is expected,
				// but []byte is more general.
				return data, nil
			}
		}
	}
	return nil, ErrResourceNotFound
}

// Resources registers a provider function for a named group of resources.
func (s *Server) Resources(name string, provider func(context.Context) ([]Resource, error)) {
	if s.resourceProviders == nil {
		s.resourceProviders = make(map[string]func(context.Context) ([]Resource, error))
	}
	s.resourceProviders[name] = provider
}

// handleResourcesList is the JSON-RPC handler for "resources/list".
// It collects all resources from registered providers.
// The router expects func(ctx *Context, params json.RawMessage) (interface{}, error)
// We made an adapter, so this original signature can be kept.
func (s *Server) handleResourcesList(stdCtx context.Context) (any, error) {
	allResources := []Resource{} // Or a struct specifically for listing
	for _, provider := range s.resourceProviders {
		res, err := provider(stdCtx)
		if err != nil {
			// TODO: Proper logging instead of just skipping
			// For now, we skip resources from a failing provider
			continue
		}
		allResources = append(allResources, res...)
	}
	return allResources, nil
}

// NOTE: This file would also need to register "resources/list" with a router,
// e.g., s.router.Handle("resources/list", s.handleResourcesList)
// That part will be handled in a subsequent step or by a different part of the SDK.
