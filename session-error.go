package skeleton

type SessionError struct {
	msg string
	err error
}

func (s *SessionError) Error() string {
	return s.msg
}

func (s *SessionError) Unwrap() error {
	return s.err
}

func NewSessionError(msg string, err error) *SessionError {
	return &SessionError{msg: msg, err: err}
}
