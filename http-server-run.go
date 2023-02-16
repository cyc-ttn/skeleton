package skeleton

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

// Runner is an instance which can be passed into Run.
type Runner interface {
	// Run should run the instance. Any functions that need to be called on
	// shutdown need to be provided here.
	Run(onShutdown ...func()) error

	// Shutdown should shutdown operation.
	Shutdown(context.Context) error
}

type RunRoutine func(<-chan struct{})

// RunDelegate is a helper which modifies the `Run` function.
type RunDelegate interface {
	// Routines are any routines that need to be run in parallel to the server.
	// For example, this could be a cleanup function or a heartbeat.
	Routines() []RunRoutine

	// WrapRun will wrap the run function, but return the error. The function
	// passed in will run the runner.Run function. For example, this could be
	// a simple logging routine.
	WrapRun(func() error) error

	// WrapShutdown should wrap the shutdown function. The function passed in
	// will run the runner.Shutdown function. For example, this could be a
	// simple logging routine.
	WrapShutdown(func() error) error
}

// Run is a helper method for the main function. It allows the user to
// dictate service initialization (e.g., DB, FileSystem, Logging), and provides
// a way for the user to define:
// - parallel routines
// - what happens around the `Run`
// - what happens around the `Shutdown`.
//
// It will also connect the parallel routines appropriately with the shutdown
// of the server, so that server shutdown also instructs these services to
// shut down.
//
// Finally, it handles the CTRL+C signal from the OS to instruct the server
// to shut down.
//
// While the exemplary usage relates to an HTTP server, any server which
// satisfies the Runner interface can be used.
//
// Usage:
// ```
// package main
//
//	func main() {
//			services := InitializeServices()
//			server := skeleton.NewLoggingHttpServer(...)
//	     	skeleton.Run(server, &CustomDelegate{})
//	}
//
// ```
func Run(runner Runner, runDelegate RunDelegate) {
	// This is for registering server shut down and shutting down the goroutines
	// that need to be shut down when the server abruptly closes. The result ot
	// the context's `Done()` function can be passed into any goroutine as a
	// case in a select statement, which should indicate shutdown.
	//
	ctxRegisterShutdown, cancel := context.WithCancel(context.Background())

	// A WaitGroup to wait for any goroutines. All of those goroutines should
	// use the context above.
	wg := sync.WaitGroup{}

	// runRoutine can be used for any routines that are required to be run.
	for _, routine := range runDelegate.Routines() {
		wg.Add(1)
		go func(fn RunRoutine) {
			defer wg.Done()
			fn(ctxRegisterShutdown.Done())
		}(routine)
	}

	// Run the server
	go func() {
		if err := runDelegate.WrapRun(func() error {
			return runner.Run(cancel)
		}); err != nil && err != http.ErrServerClosed {
			os.Exit(2)
		}
	}()

	// Handle interrupt signal
	cInterrupt := make(chan os.Signal, 1)
	signal.Notify(cInterrupt, os.Interrupt, os.Kill)
	<-cInterrupt

	ctxShutdown, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_ = runDelegate.WrapShutdown(func() error {
		if err := runner.Shutdown(ctxShutdown); err != nil {
			return err
		}
		wg.Wait()
		return nil
	})

}
