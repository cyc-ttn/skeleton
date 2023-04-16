package skeleton

import (
	"errors"
	"net/http"

	"cloud.google.com/go/logging"
	"github.com/cyc-ttn/gorouter"
	"github.com/monstercat/golib/logger"
)

var (
	ErrInvalidRoute = errors.New("invalid route")
)

// GoRouterRoute wraps gorouter.Route[R] so that it includes a
// gorouter.RouteContext object.
type GoRouterRoute[R any] struct {
	*gorouter.RouteContext
	gorouter.Route[R]
}

// wrapGoRouter is a special struct that wraps gorouter, so that it properly
// implements Router.
type wrapGoRouter[Ctx any] struct {
	*gorouter.RouterNode[Ctx]
}

// AddRoute adds a route to the router. The AddRoute function here requires
// that the route implement the gorouter.Route interface. otherwise,
// ErrInvalidRoute will be returned.
func (r *wrapGoRouter[Ctx]) AddRoute(route Route[Ctx]) error {
	rV, ok := route.(gorouter.Route[Ctx])
	if !ok {
		return ErrInvalidRoute
	}
	return r.RouterNode.AddRoute(rV)
}

// Match should match the provided method and path to a route. If nil is
// returned, a NotFound error will automatically be returned by the
// HttpServer. This Route object should also include any matches that could
// be desired from parsing the path. For example, if the router allows
// route patterns with placeholders such as :id, the matching ID can be
// provided within this returned R
func (r *wrapGoRouter[Ctx]) Match(method, path string) (*GoRouterRoute[Ctx], error) {
	ctx := &gorouter.RouteContext{}
	route, err := r.RouterNode.Match(method, path, ctx)
	if err != nil {
		return nil, err
	}
	return &GoRouterRoute[Ctx]{
		Route:        route,
		RouteContext: ctx,
	}, nil
}

// GoRouter providers a Router which can be used in skeleton.HttpServer or
// skeleton.LoggingHttpServer.
func GoRouter[Ctx any]() Router[Ctx, *GoRouterRoute[Ctx]] {
	return &wrapGoRouter[Ctx]{
		RouterNode: gorouter.NewRouter[Ctx](),
	}
}

// GoRoute creates a skeleton.Route whose underlying implementation is a
// gorouter.DefaultRoute.
func GoRoute[Ctx any](method string, path string, fn func(ctx Ctx)) Route[Ctx] {
	return &gorouter.DefaultRoute[Ctx]{
		Method:      method,
		Path:        path,
		HandlerFunc: fn,
	}
}

// GoHttpServerDelegate is a server delegate that returns a
// gorouter.RouteContext. Note that the base gorouter.RouteContext is meant
// to be encapsulated in another struct which is used to provide Session
// functionality. As a result, this delegate also does not use the session
// variable.
//
// Please encapsulate this delegate to provide session related functionality
// to the internal variables.
type GoHttpServerDelegate struct{}

// Generate should generate a context to pass into the routes. The route
// and related session is provided. This is used only in the Serve and
// ServeHTTP functions.
func (d *GoHttpServerDelegate) Generate(
	w http.ResponseWriter,
	req *http.Request,
	r *GoRouterRoute[*gorouter.RouteContext],
	s Session,
) *gorouter.RouteContext {
	return &gorouter.RouteContext{
		W:      w,
		R:      req,
		Method: r.Method,
		Path:   r.Path,
		Params: r.Params,
		Query:  r.Query,
	}
}

type LoggingGoHttpServerDelegate struct {
	GoHttpServerDelegate
}

// RequestLogger generates a logger specific to the HTTP request.
func (d *LoggingGoHttpServerDelegate) RequestLogger(l logger.Logger, r *http.Request) logger.HTTPRequest {
	// Request for Google Logger
	req := &logging.HTTPRequest{
		Request: r,
	}

	return &logger.GoogleHTTPRequest{
		Logger:     l,
		LogRequest: req,
	}
}

// Generate generates a context to pass into the routes. The route, related
// session and base logger is provided. Note that this also ignores the loggers
// as the standard gorouter.RouteContext does not allow for routers.
func (d *LoggingGoHttpServerDelegate) Generate(
	w http.ResponseWriter,
	req *http.Request,
	r *GoRouterRoute[*gorouter.RouteContext],
	s Session,
	l logger.Logger,
	lr logger.HTTPRequest,
) *gorouter.RouteContext {
	return d.GoHttpServerDelegate.Generate(w, req, r, s)
}
