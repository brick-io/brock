package sdk

import (
	"errors"
)

var _ ErrorWrapper = (*WrapError)(nil)

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
