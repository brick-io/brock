package brock

import (
	"fmt"
)

type (
	// print.

	State      = fmt.State
	Formatter  = fmt.Formatter
	Stringer   = fmt.Stringer
	GoStringer = fmt.GoStringer

	// scan.

	ScanState = fmt.ScanState
	Scanner   = fmt.Scanner
)

// nolint: gochecknoglobals
var (
	NonNil = struct{}{}

	Errorf = fmt.Errorf

	// print.

	Fprintf = fmt.Fprintf
	Printf  = fmt.Printf
	Sprintf = fmt.Sprintf

	Fprint = fmt.Fprint
	Print  = fmt.Print
	Sprint = fmt.Sprint

	Fprintln = fmt.Fprintln
	Println  = fmt.Println
	Sprintln = fmt.Sprintln

	// scan.
	Scan   = fmt.Scan
	Scanln = fmt.Scanln
	Scanf  = fmt.Scanf

	Sscan   = fmt.Sscan
	Sscanln = fmt.Sscanln
	Sscanf  = fmt.Sscanf

	Fscan   = fmt.Fscan
	Fscanln = fmt.Fscanln
	Fscanf  = fmt.Fscanf
)

func Apply[T any](opt T, opts ...func(T)) T {
	for _, fn := range opts {
		if fn != nil {
			fn(opt)
		}
	}
	return opt
}

func Yield[T any](v T) func() T {
	return func() T { return v }
}

func IfThenElse[T any](cond bool, this, that T) T {
	if cond {
		return this
	}
	return that
}

func Must[T any](v T, err error) T {
	PanicIf(err != nil, v)
	return v
}

func PanicIf(cond bool, v any) {
	if cond {
		panic(v)
	}
}

// RecoverPanic is a callback func to recover from panic
//
//	defer RecoverPanic(func(v any) { print(v) })()
func RecoverPanic(fn func(v any)) func() {
	return func() {
		if fn != nil {
			fn(recover())
		}
	}
}
