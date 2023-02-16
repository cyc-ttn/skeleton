package skeleton

import (
	"net/http"

	"github.com/monstercat/golib/logger"
)

// LoggingHttpServerDelegate describes how certain parts of the
// LoggingHttpServer should run. In other words, the LoggingHttpServer
// delegates this functionality to the struct that implements the
// LoggingHttpServerDelegate interface.
type LoggingHttpServerDelegate[Ctx any, R Route[Ctx]] interface {
	// RequestLogger generates a logger specific to the HTTP request.
	RequestLogger(l logger.Logger, r *http.Request) logger.HTTPRequest

	// Generate generates a context to pass into the routes. The route, related
	// session and base logger is provided.
	Generate(R, Session, logger.Logger, logger.HTTPRequest) Ctx
}

// HttpServerDelegateBridge bridges between an HttpServerDelegate and a
// LoggingServerDelegate. It implements HttpServerDelegate, and requires a
// LoggingHttpServerDelegate.
type HttpServerDelegateBridge[Ctx any, R Route[Ctx]] struct {
	Logger        logger.Logger
	RequestLogger logger.HTTPRequest
	Delegate      LoggingHttpServerDelegate[Ctx, R]
}

func NewHttpServerDelegateBridge[Ctx any, R Route[Ctx]](l logger.Logger, req logger.HTTPRequest) *HttpServerDelegateBridge[Ctx, R] {
	return &HttpServerDelegateBridge[Ctx, R]{
		Logger:        l,
		RequestLogger: req,
	}
}

func (b *HttpServerDelegateBridge[Ctx, R]) Generate(r R, sess Session) Ctx {
	return b.Delegate.Generate(r, sess, b.Logger, b.RequestLogger)
}
