package skeleton

import (
	"errors"

	"github.com/cyc-ttn/gorouter"
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

// GetHandler implements the Route[R] interface
func (w *GoRouterRoute[R]) GetHandler() HandlerFunc[R] {
	return HandlerFunc[R](w.Route.GetHandler())
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

// GoRouter providers a Router which can be
func GoRouter[Ctx any]() Router[Ctx, *GoRouterRoute[Ctx]] {
	return &wrapGoRouter[Ctx]{
		RouterNode: gorouter.NewRouter[Ctx](),
	}
}
