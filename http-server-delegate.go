package skeleton

import "net/http"

type HttpServerDelegate[Ctx any, R Route[Ctx]] interface {
	// Generate should generate a context to pass into the routes. The route
	// and related session is provided. This is used only in the Serve and
	// ServeHTTP functions.
	Generate(http.ResponseWriter, *http.Request, R, Session) Ctx
}

// HttpServerDelegateFunc implements HttpServerDelegate based on a provided
// function.
type HttpServerDelegateFunc[Ctx any, R Route[Ctx]] func(http.ResponseWriter, *http.Request, R, Session) Ctx

func (f HttpServerDelegateFunc[Ctx, R]) Generate(w http.ResponseWriter, req *http.Request, r R, s Session) Ctx {
	return f(w, req, r, s)
}
