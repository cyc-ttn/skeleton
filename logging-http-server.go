package skeleton

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/monstercat/golib/logger"
)

// LoggingHttpServer is an HTTP server that includes a logging mechanism. Use
// NewLoggingHttpServer to initialize this server. Internally, this server
// includes a version of HttpServer without its delegate, and calls
// HttpServer.ServeWithDelegate.
type LoggingHttpServer[Ctx any, R Route[Ctx]] struct {
	logger.Logger
	*HttpServer[Ctx, R]

	// Delegate instructs the LoggingHttpServer how to create a request logger
	// and how to generate the context, including the logger.
	//
	// The Request Logger needs to be of type logger.HTTPRequest
	Delegate LoggingHttpServerDelegate[Ctx, R]
}

// NewLoggingHttpServer creates a new HTTP server with logging capability.
// User can decide what Router, SessionStore, and context to provide all route
// handlers.
//
// All route handlers are of the form func(Ctx), where Ctx can be anything.
// We will refer to that "Ctx" as the RouteContext.
//
// The RouteContext is created via the LoggingHttpServerDelegate.Generate
// method. To pass in extra variables such as services to the RouteContext,
// provide it to a custom delegate struct implementing HttpServerDelegate and
// return it in the LoggingHttpServerDelegate.Generate function.
func NewLoggingHttpServer[Ctx any, R Route[Ctx]](
	l logger.Logger,
	S SessionStore,
	Router Router[Ctx, R],
	Delegate LoggingHttpServerDelegate[Ctx, R],
) *LoggingHttpServer[Ctx, R] {
	server := &LoggingHttpServer[Ctx, R]{
		Logger: l,
		HttpServer: &HttpServer[Ctx, R]{
			S: S,
			R: Router,
		},
		Delegate: Delegate,
	}
	return server
}

// Run the server. This is a blocking function.
func (s *LoggingHttpServer[Ctx, R]) Run(onShutdown ...func()) error {
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

// ServeHTTP allows LoggingHttpServer to implement the http.Handler interface.
func (s *LoggingHttpServer[Ctx, R]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestId := uuid.New().String() // Request ID (unique to the current request)
	w.Header().Set("request-id", requestId)

	// Generate the request logger.
	reqLogger := s.Delegate.RequestLogger(s, r)

	// Initialize the timer.
	reqLogger.StartTimer()

	// generate a logger for this specific request. This would be a
	// contextual logger wrapping an HTTP logger.
	lgr := &logger.Contextual{
		Context: logger.NewContext("Request", map[string]interface{}{
			"ID":     requestId,
			"Method": r.Method,
			"Path":   r.URL.Path,
		}),
		Logger: reqLogger,
	}

	// Serve based on the route. We need to pass in a special delegate (since
	// the HttpServer's delegate is nil.
	err := s.HttpServer.ServeWithDelegate(w, r, NewHttpServerDelegateBridge[Ctx, R](lgr, reqLogger))
	if err == nil {
		return
	}
	if err == ErrNoRoute {
		lgr.Log(logger.SeverityWarning, "Could not find route")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("The system services are temporarily unavailable at the moment."))
	reqLogger.Log(logger.SeverityError, err)
	return
}
