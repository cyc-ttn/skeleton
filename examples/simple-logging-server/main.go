package main

import (
	"net/http"

	"github.com/cyc-ttn/gorouter"
	"github.com/monstercat/golib/logger"

	"github.com/cyc-ttn/skeleton"
)

func main() {
	// A logger that logs to standard output.
	l := &logger.Standard{}

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

	// The provided delegate which works with skeleton.GoRouter and
	// gorouter.RouteContext. For most scenarios, a custom one should be
	// provided.
	//
	// Note that this delegate *actually* does not make use of the provided
	// loggers. It is provided merely as an example. A custom delegate should
	// be created matching the logging needs of the application. Two logging
	// utilities are provided in the `Generate` function.
	//
	// 1. A contextualized logger for normal logging situations
	// 2. A request logger which includes helpful methods for storing the
	//    HTTP response status, or whether the response was cached. It also
	//    allows functions for starting a request timer and setting that the
	//    request had finished. Of course, the implementation for the request
	//    logger is customizable by the user.
	//
	// The first is the  logger passed into the HttpServer
	// initialization below, and the second  is a logger specific to the
	// request, which is generated through the
	// LoggingHttpServerDelegate.RequestLogger.
	del := &skeleton.LoggingGoHttpServerDelegate{}

	// Create the HTTP server. We are creating one without a session,
	s := skeleton.NewLoggingHttpServer[*gorouter.RouteContext, *skeleton.GoRouterRoute[*gorouter.RouteContext]](
		l,
		":80",
		nil,
		router,
		del,
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
