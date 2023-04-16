package skeleton

// NilRunDelegate is a run delegate that does nothing.
type NilRunDelegate struct {
}

// Routines are any routines that need to be run in parallel to the server.
// For example, this could be a cleanup function or a heartbeat.
func (d *NilRunDelegate) Routines() []RunRoutine {
	return nil
}

// WrapRun will wrap the run function, but return the error. The function
// passed in will run the runner.Run function. For example, this could be
// a simple logging routine.
func (d *NilRunDelegate) WrapRun(fn func() error) error {
	return fn()
}

// WrapShutdown should wrap the shutdown function. The function passed in
// will run the runner.Shutdown function. For example, this could be a
// simple logging routine.
func (d *NilRunDelegate) WrapShutdown(fn func() error) error {
	return fn()
}
