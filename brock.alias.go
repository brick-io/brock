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

// Nop is not really doing anything,
// just there for us not to use the underscore
//
//	Nop(a, err)
func Nop(...any) {
	// Nop is not really doing anything,
	// just there for us not to use the underscore
	//  Nop(a, err)
}

func Ref[T any](v T) *T { return &v }

func Val[T any](v *T) T { return IfThenElse(v == nil, *new(T), *v) }

func Yield[T any](v T) func() T { return func() T { return v } }

func Apply[T any](opt T, opts ...func(T)) T {
	for _, fn := range opts {
		if fn != nil {
			fn(opt)
		}
	}
	return opt
}

func IfThenElse[T any](cond bool, this, that T) T {
	if cond {
		return this
	}
	return that
}
