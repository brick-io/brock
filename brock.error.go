package brock

var (
	ErrAlreadySent        = Errorf("brock: http: already sent to the client")
	ErrAlreadyStreamed    = Errorf("brock: http: already streamed to the client")
	ErrRequestCancelled   = Errorf("brock: http: request cancelled")
	ErrEmptyResponse      = Errorf("brock: http: empty response")
	ErrInvalidArguments   = Errorf("brock: sql: invalid arguments for scan")
	ErrInvalidTransaction = Errorf("brock: sql: invalid transaction")
	ErrNoColumns          = Errorf("brock: sql: no columns returned")
	ErrUnimplemented      = Errorf("brock: unimplemented")
)

type AssertionError struct {
	A any
	M string
	E any
}

func (err *AssertionError) Error() string {
	return Sprintf("brock: assert: [actual][%v] is not %s with [expected][%v]")
}

func (err *AssertionError) With(matcher string, actual, expected any) error {
	err.A = actual
	err.M = matcher
	err.E = expected
	return err
}

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
	s := ""
	if err.Total > 1 {
		s = "s"
	}
	return Sprintf("brock: sql: Unable to connect to database on index %d with total %d element%s.",
		err.Index,
		err.Total,
		s,
	)
}

func Errors(errs ...error) error {
	var err error
	for _, e := range errs {
		if e == nil {
			continue
		}
		if err == nil {
			err = e
			continue
		}

		err = Errorf("%w -> [%s]", err, e)
	}
	return err
}
