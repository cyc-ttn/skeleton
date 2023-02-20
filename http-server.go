package skeleton

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrNoRoute = errors.New("could not find route")
)

// Route defines what should be returned by the router. If the desired route
// does not comply with the current interface, wrap the route returned from
// the desired router to comply.
type Route[R any] interface {
	GetHandler() func(R)
}

// Router defines the methods required by HttpServer for route handler
// organization. If the desired router does not fit this mold, simply wrap it
// in another struct to force it to comply.
//
// ```
//
//	type CustomRouter struct {
//	   *TheRouterIWant
//	}
//
//	func (r *CustomRouter) AddRoute(route Route[R]) error {
//	   // Do what is needed here to add the route to the desired router
//	}
//
// ```
type Router[Ctx any, R Route[Ctx]] interface {
	// AddRoute adds a route to the router. Note that the added route does *not*
	// need to be the same as R.
	AddRoute(route Route[Ctx]) error

	// Match should match the provided method and path to a route. If nil is
	// returned, a NotFound error will automatically be returned by the
	// HttpServer. This Route object should also include any matches that could
	// be desired from parsing the path. For example, if the router allows
	// route patterns with placeholders such as :id, the matching ID can be
	// provided within this returned R
	Match(method, path string) (R, error)
}

// HttpServer describes an extensible and basic http server implementation.
// User can decide what Router, SessionStore, and context to provide all route
// handlers.
//
// All route handlers are of the form func(Ctx), where Ctx can be anything.
// We will refer to that "Ctx" as the RouteContext.
//
// The RouteContext is created via the HttpServerDelegate.Generate method. To
// pass in extra variables such as services to the RouteContext, provide it
// to a custom delegate struct implementing HttpServerDelegate and return it
// in the HttpServerDelegate.Generate function.
type HttpServer[Ctx any, R Route[Ctx]] struct {
	Addr   string
	Server *http.Server   // The base http Server.
	S      SessionStore   // Storage for session information
	R      Router[Ctx, R] // Router for organization route handlers.

	// Delegate should be provided by the application
	Delegate HttpServerDelegate[Ctx, R]
}

// NewHttpServer creates a new HTTP server.
func NewHttpServer[Ctx any, R Route[Ctx]](
	addr string,
	S SessionStore,
	Router Router[Ctx, R],
	Delegate HttpServerDelegate[Ctx, R],
) *HttpServer[Ctx, R] {
	return &HttpServer[Ctx, R]{
		Addr:     addr,
		S:        S,
		R:        Router,
		Delegate: Delegate,
	}
}

// Run the server. This is a blocking function.
func (s *HttpServer[Ctx, R]) Run(onShutdown ...func()) error {
	defer s.S.Shutdown()

	// This is so that we can handle cleanup
	s.Server = &http.Server{
		Addr:    s.Addr,
		Handler: s,
	}
	for _, f := range onShutdown {
		s.Server.RegisterOnShutdown(f)
	}
	return s.Server.ListenAndServe()
}

// Shutdown the server,
func (s *HttpServer[Ctx, R]) Shutdown(ctx context.Context) error {
	if s.Server == nil {
		return nil
	}
	return s.Server.Shutdown(ctx)
}

// ServeWithDelegate is a version of serve allowing for a custom delegate to be
// provided.
func (s *HttpServer[Ctx, R]) ServeWithDelegate(w http.ResponseWriter, r *http.Request, delegate HttpServerDelegate[Ctx, R]) error {
	// Initialize the session. If there is an error, return the default error
	// message.
	sess, err := s.S.Get(r)
	if err != nil {
		return NewSessionError("unable to get session", err)
	}

	// Retrieve a route, if possible.
	route, err := s.R.Match(r.Method, r.URL.Path)
	if err != nil {
		return ErrNoRoute
	}

	// Generate the context. It is assumed here that Generator is provided, as
	// it is required.
	ctx := delegate.Generate(route, sess)
	route.GetHandler()(ctx)
	return nil
}

// Serve is a version of ServeHTTP which returns an error. This is useful for
// applications which want to extend the basic functionality of the HttpServer
// while using its default implementation.If the ServeHTTP functionality is
// acceptable, application can use that as well.
//
// The basic functionality includes getting a session object, retrieving the
// route, and calling the handler on the route.
//
// In the case that there is an error in retrieving the session, a SessionError
// object will be returned. Use SessionError.Unwrap to view the underlying error.
//
// In the case that a route could not be found, ErrNoRoute will be returned.
func (s *HttpServer[Ctx, R]) Serve(w http.ResponseWriter, r *http.Request) error {
	return s.ServeWithDelegate(w, r, s.Delegate)
}

// ServeHTTP implements the http.Handler interface.
func (s *HttpServer[Ctx, R]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := s.Serve(w, r)
	if err == nil {
		return
	}

	if err == ErrNoRoute {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("The system services are temporarily unavailable at the moment."))
}
