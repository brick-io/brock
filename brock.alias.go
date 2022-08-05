package brock

import (
	"errors"
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

//nolint: gochecknoglobals
var (
	Errorf = fmt.Errorf
	Error  = errors.New

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
