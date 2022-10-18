package brock

import (
	"errors"
)

var (
	ErrCryptoInvalidPEMFormat    = Errorf("brock: crypto: invalid pem format")
	ErrCryptoInvalidKeypair      = Errorf("brock: crypto: invalid keypair")
	ErrCryptoUnsupportedKeyTypes = Errorf("brock: crypto: unsupported key type")
	ErrFSMEmptyStates            = Errorf("brock: fsm: empty states")
	ErrFSMNoInitialStates        = Errorf("brock: fsm: no initial states")
	ErrHTTPAlreadySent           = Errorf("brock: http: already sent to the client")
	ErrHTTPAlreadyStreamed       = Errorf("brock: http: already streamed to the client")
	ErrHTTPRequestCancelled      = Errorf("brock: http: request cancelled")
	ErrHTTPEmptyResponse         = Errorf("brock: http: empty response")
	ErrSQLInvalidArguments       = Errorf("brock: sql: invalid arguments for scan")
	ErrSQLInvalidCommand         = Errorf("brock: sql: invalid command")
	ErrSQLInvalidTransaction     = Errorf("brock: sql: invalid transaction")
	ErrSQLMultipleCommands       = Errorf("brock: sql: multiple commands")
	ErrSQLNoColumns              = Errorf("brock: sql: no columns returned")
	ErrUnimplemented             = Errorf("brock: unimplemented")

	_ ErrorWrapper = (*WrapError)(nil)
)

type ErrorWrapper interface {
	error
	Unwrap() error
	As(target any) bool
	Is(target error) bool
}

type WrapError struct {
	Err error
	Msg string

	// default error string
	Redact string
}

func (err *WrapError) Error() string {
	switch {
	case err.Redact != "":
		return err.Redact
	case err.Err == nil:
		return err.Msg
	case err.Msg == "":
		return err.Err.Error()
	}

	return err.Msg + ": " + err.Err.Error()
}

func (err *WrapError) Unwrap() error { return err.Err }

func (err *WrapError) As(target any) bool { return errors.As(err.Err, target) }

func (err *WrapError) Is(target error) bool { return errors.Is(err.Err, target) }

func (err *WrapError) MarshalJSON() ([]byte, error) { return JSON.Marshal(err.Error()) }

type SQLMismatchColumnsError struct{ Col, Dst int }

func (err *SQLMismatchColumnsError) Error() string {
	return Sprintf("brock: sql: mismatch %d columns on %d destinations",
		err.Col,
		err.Dst,
	)
}

// SQLRoundRobinError reporting issue when getting from set of Conn from SQLRoundRobin.
type SQLRoundRobinError struct{ Total, Index int }

func (err *SQLRoundRobinError) Error() string {
	return Sprintf("brock: sql: Unable to connect to database on index %d with total %d element%s.",
		err.Index,
		err.Total,
		IfThenElse(err.Total > 1, "s", ""),
	)
}

// Errors combine multiple errors into one error.
type Errors []error

func (errs Errors) Error() string {
	out := make([]string, 0)

	for _, each := range errs {
		if each != nil {
			out = append(out, each.Error())
		} else {
			out = append(out, "")
		}
	}

	p, _ := JSON.Marshal(out)

	return IfThenElse(len(out) < 1, "", string(p))
}

func (errs Errors) Filter(fn func(error) bool) Errors {
	out := make(Errors, 0)

	for _, each := range errs {
		if fn(each) {
			out = append(out, each)
		}
	}

	return IfThenElse(len(out) < 1, nil, out)
}
