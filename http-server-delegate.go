package skeleton

type HttpServerDelegate[Ctx any, R Route[Ctx]] interface {
	// Generate should generate a context to pass into the routes. The route
	// and related session is provided. This is used only in the Serve and
	// ServeHTTP functions.
	Generate(R, Session) Ctx
}

// HttpServerDelegateFunc implements HttpServerDelegate based on a provided
// function.
type HttpServerDelegateFunc[Ctx any, R Route[Ctx]] func(R, Session) Ctx

func (f HttpServerDelegateFunc[Ctx, R]) Generate(r R, s Session) Ctx {
	return f(r, s)
}
