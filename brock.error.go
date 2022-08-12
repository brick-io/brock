package brock

import (
	"errors"
	"strings"
)

var (
	ErrAlreadySent        = Errorf("brock: http: already sent to the client")
	ErrAlreadyStreamed    = Errorf("brock: http: already streamed to the client")
	ErrRequestCancelled   = Errorf("brock: http: request cancelled")
	ErrEmptyResponse      = Errorf("brock: http: empty response")
	ErrInvalidArguments   = Errorf("brock: sql: invalid arguments for scan")
	ErrInvalidCommand     = Errorf("brock: sql: invalid command")
	ErrInvalidTransaction = Errorf("brock: sql: invalid transaction")
	ErrMultipleCommands   = Errorf("brock: sql: multiple commands")
	ErrNoColumns          = Errorf("brock: sql: no columns returned")
	ErrUnimplemented      = Errorf("brock: unimplemented")

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
}

func (err *WrapError) Error() string { return IfThenElse(err.Msg != "", err.Msg, err.Err.Error()) }

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

type Errors []error

// Errors combine multiple errors into one error
func (errs Errors) Error() string {
	out := make([]string, 0)
	for _, each := range errs {
		if each == nil {
			continue
		}

		out = append(out, each.Error())
	}
	return strings.Join(out, ", ")
}
