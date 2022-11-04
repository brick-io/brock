package sdksql

import (
	"context"
	"database/sql"
	"sync"

	"github.com/brick-io/brock/sdk"
)

var (
	ErrInvalidCommand   = sdk.Errorf("brock/sdksql: invalid command")
	ErrMultipleCommands = sdk.Errorf("brock/sdksql: multiple commands")
)

// RoundRobinError reporting issue when getting from set of Conn from RoundRobin.
type RoundRobinError struct{ Total, Index int }

func (err *RoundRobinError) Error() string {
	return sdk.Sprintf("brock/sdksql: Unable to connect to database on index %d with total %d element%s.",
		err.Index,
		err.Total,
		sdk.IfThenElse(err.Total > 1, "s", ""),
	)
}

// =============================================================================

func RoundRobin(conns ...Conn) Conn {
	conns2 := make([]Conn, 0)

	for _, v := range conns {
		if v != nil {
			conns2 = append(conns2, v)
		}
	}

	return &roundrobin{conns2, 0, new(sync.Mutex)}
}

type roundrobin struct {
	conns []Conn
	index int
	mutex *sync.Mutex
}

// conn will return a new Conn that balanced using roundRobin
//
//	rr.conn(0)    -> direct READ+WRITE
//	rr.conn(1..n) -> direct READ-ONLY
//	rr.conn(-1)   -> roundRobin READ+WRITE and READ-ONLY
//	rr.conn(-2)   -> roundRobin READ-ONLY
func (x *roundrobin) conn(i int) (Conn, error) {
	l := len(x.conns)

	switch {
	case l == 1: // only one
		x.index = 0
	case i >= 0 && l > i: // direct
		x.index = i
	case (i == -1 || i == -2) && l > 1: // roundRobin
		x.mutex.Lock()
		if x.index++; x.index >= l {
			switch i {
			case -1: // roundRobin READ/WRITE and READ-ONLY
				x.index = 0
			case -2: // roundRobin READ-ONLY
				x.index = 1
			}
		}
		x.mutex.Unlock()
	default:
		return nil, &RoundRobinError{l, i}
	}

	return x.conns[x.index], nil
}

func (x *roundrobin) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	conn, err := x.conn(0)
	if err != nil {
		return nil, err
	}

	return conn.BeginTx(ctx, opts)
}

// Close all databases.
func (x *roundrobin) Close() error {
	errs := make(sdk.Errors, 0)
	for i := range x.conns {
		errs = append(errs, x.conns[i].Close())
	}

	return errs
}

// PingContext all databases.
func (x *roundrobin) PingContext(ctx context.Context) error {
	errs := make(sdk.Errors, 0)
	for i := range x.conns {
		errs = append(errs, x.conns[i].PingContext(ctx))
	}

	return errs
}

const format = "%w: %q"

// PrepareContext valid queries are DDL, DML & SELECT.
func (x *roundrobin) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	query = Tool.RemoveComment(query)

	if Tool.IsMultipleCommand(query) {
		return nil, ErrMultipleCommands
	} else if !Tool.IsValidCommand(query) {
		return nil, sdk.Errorf(format, ErrInvalidCommand, query)
	}

	conn, err := Conn(nil), error(nil)
	if Tool.IsDDLCommand(query) {
		conn, err = x.conn(0)
	} else if Tool.IsDMLCommand(query) {
		conn, err = x.conn(0)
	} else if Tool.IsSELECTCommand(query) {
		conn, err = x.conn(-2)
	} else {
		return nil, sdk.Errorf(format, ErrInvalidCommand, query)
	}

	if err != nil {
		return nil, err
	}

	return conn.PrepareContext(ctx, query)
}

// ExecContext valid queries are DDL & DML.
func (x *roundrobin) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	query = Tool.RemoveComment(query)

	if Tool.IsMultipleCommand(query) {
		return nil, ErrMultipleCommands
	} else if !Tool.IsValidCommand(query) {
		return nil, sdk.Errorf(format, ErrInvalidCommand, query)
	}

	conn, err := Conn(nil), error(nil)
	if Tool.IsDDLCommand(query) {
		conn, err = x.conn(0)
	} else if Tool.IsDMLCommand(query) {
		conn, err = x.conn(0)
	} else if Tool.IsSELECTCommand(query) {
		return nil, sdk.Errorf(format, ErrInvalidCommand, query)
	} else {
		return nil, sdk.Errorf(format, ErrInvalidCommand, query)
	}

	if err != nil {
		return nil, err
	}

	return conn.ExecContext(ctx, query, args...)
}

// QueryContext valid queries are SELECT.
func (x *roundrobin) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	query = Tool.RemoveComment(query)

	if Tool.IsMultipleCommand(query) {
		return nil, ErrMultipleCommands
	} else if !Tool.IsValidCommand(query) {
		return nil, sdk.Errorf(format, ErrInvalidCommand, query)
	}

	conn, err := Conn(nil), error(nil)
	if Tool.IsSELECTCommand(query) {
		conn, err = x.conn(-2)
	} else if Tool.IsDDLCommand(query) {
		return nil, sdk.Errorf(format, ErrInvalidCommand, query)
	} else if Tool.IsDMLCommand(query) {
		return nil, sdk.Errorf(format, ErrInvalidCommand, query)
	} else {
		return nil, sdk.Errorf(format, ErrInvalidCommand, query)
	}

	if err != nil {
		return nil, err
	}

	return conn.QueryContext(ctx, query, args...)
}

// QueryRowContext valid queries are SELECT.
func (x *roundrobin) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	query = Tool.RemoveComment(query)

	if Tool.IsMultipleCommand(query) {
		return nil
	} else if !Tool.IsValidCommand(query) {
		return nil
	}

	conn, err := Conn(nil), error(nil)
	if Tool.IsSELECTCommand(query) {
		conn, err = x.conn(-2)
	} else if Tool.IsDDLCommand(query) {
		return nil
	} else if Tool.IsDMLCommand(query) {
		return nil
	} else {
		return nil
	}

	if err != nil {
		return nil
	}

	return conn.QueryRowContext(ctx, query, args...)
}
