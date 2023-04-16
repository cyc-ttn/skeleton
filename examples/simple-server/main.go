package main

import (
	"net/http"

	"github.com/cyc-ttn/gorouter"

	"github.com/cyc-ttn/skeleton"
)

// This example creates a simple HTTP server with a single route and no session
// storage. It uses default provided routers and
func main() {

	// A wrapper around http://github.com/cyc-ttn/gorouter. We are using this
	// as the default router. Feel free to use any router that satisfies
	// skeleton.Router. If the intended route does not satisfy, encapsulate
	// the router to cause it to satisfy.
	router := skeleton.GoRouter[*gorouter.RouteContext]()

	// Add a route for GET /. It simply returns OK.
	router.AddRoute(skeleton.GoRoute[*gorouter.RouteContext](
		http.MethodGet,
		"/",
		func(ctx *gorouter.RouteContext) {
			ctx.W.WriteHeader(http.StatusOK)
			ctx.W.Write([]byte("OK"))
		},
	))

	// Create the HTTP server. We are creating one without a session, and using
	// the provided delegate which works with skeleton.GoRouter and
	// gorouter.RouteContext. For most scenarios, a custom one should be
	// provided.
	s := skeleton.NewHttpServer[*gorouter.RouteContext, *skeleton.GoRouterRoute[*gorouter.RouteContext]](
		":80",
		nil,
		router,
		&skeleton.GoHttpServerDelegate{},
	)

	// Run the HTTP server. The RunDelegate can be used to insert any routines
	// that need to be run. If no routines need to be run, simply set nil.
	//
	// A routine is generally a CRON-like process. For example, a process to
	// clean-up the local cache. It takes in a channel which is closed when
	// the goroutine should end. The channel will be closed, for example, when
	// a signal to interrupt the current HTTP server is detected.
	skeleton.Run(s, nil)
}
