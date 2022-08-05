package brock

import (
	"context"
	"database/sql"
	"io"
	"strings"
	"sync"
)

var SQL brock_sql

type brock_sql struct {
	Box    brock_sql_box
	Helper brock_sql_helper
}

type (
	BeginTx interface {
		BeginTx(ctx context.Context, opts *sql.TxOptions) (tx *sql.Tx, err error)
	}
	ExecContext interface {
		ExecContext(ctx context.Context, query string, args ...any) (res sql.Result, err error)
	}
	PingContext interface {
		PingContext(ctx context.Context) (err error)
	}
	PrepareContext interface {
		PrepareContext(ctx context.Context, query string) (stmt *sql.Stmt, err error)
	}
	QueryContext interface {
		QueryContext(ctx context.Context, query string, args ...any) (rows *sql.Rows, err error)
	}
	QueryRowContext interface {
		QueryRowContext(ctx context.Context, query string, args ...any) (row *sql.Row)
	}
)

// SQLConn is a common interface of *sql.DB and *sql.Conn.
type SQLConn interface {
	BeginTx
	io.Closer
	PingContext
	SQLTxConn
}

// SQLTxConn is a common interface of *sql.DB, *sql.Conn, and *sql.Tx.
type SQLTxConn interface {
	ExecContext
	PrepareContext
	QueryContext
	QueryRowContext
}

func (brock_sql) Discard() any { return new([]byte) }

type brock_sql_box struct{}

// =============================================================================

// Exec will wrap `ExecContext` so that we can Scan later
//  Exec(cmd.ExecContext(ctx, "..."))
func (brock_sql_box) Exec(val sql.Result, err error) SQLBoxExec {
	return brock_sql_box_exec{val, err}
}

type SQLBoxExec interface {
	// Scan the result of ExecContext that usually return numbers of rowsAffected
	// and lastInsertID.
	Scan(rowsAffected *int, lastInsertID *int) error
}

type brock_sql_box_exec struct {
	res sql.Result
	err error
}

func (x brock_sql_box_exec) Scan(rowsAffected *int, lastInsertID *int) error {
	n, err := int64(0), x.err
	if err != nil {
		return err
	}

	if x.res == nil {
		return ErrInvalidArguments
	}

	if rowsAffected != nil {
		if n, err = x.res.RowsAffected(); err != nil {
			return err
		}
		*rowsAffected = int(n)
	}

	if lastInsertID != nil {
		if n, err = x.res.LastInsertId(); err != nil {
			return err
		}
		*lastInsertID = int(n)
	}

	return err
}

// =============================================================================

// Query will wrap `QueryContext` so that we can Scan later
//  Query(cmd.QueryContext(ctx, "..."))
func (brock_sql_box) Query(val *sql.Rows, err error) SQLBoxQuery {
	return brock_sql_box_query{val, err}
}

type SQLBoxQuery interface {
	// Scan accept do, a func that accept `i int` as index and returning list
	// of pointers.
	//  pointers == nil   // break the loop
	//  len(pointers) < 1 // skip the current loop
	//  len(pointers) > 0 // assign the pointer, MUST be same as the length of columns
	Scan(row func(i int) (pointers []any)) error
}

type brock_sql_box_query struct {
	res *sql.Rows
	err error
}

func (x brock_sql_box_query) Scan(row func(i int) []any) error {
	err := x.err
	if err != nil {
		return err
	} else if x.res == nil {
		return sql.ErrNoRows
	} else if err = x.res.Err(); err != nil {
		return err
	}
	defer x.res.Close()

	cols, err := x.res.Columns()
	if err != nil {
		return err
	} else if len(cols) < 1 {
		return ErrNoColumns
	}

	for i := 0; x.res.Next(); i++ {
		err = x.res.Err()
		if err != nil {
			return err
		}

		dest := row(i)
		if dest == nil { // nil dest
			break
		} else if len(dest) < 1 { // empty dest
			continue
		} else if len(dest) != len(cols) { // diff dest & cols
			return &SQLMismatchColumnsError{len(cols), len(dest)}
		}

		err = x.res.Scan(dest...) // scan into pointers
		if err != nil {
			return err
		}
	}

	return err
}

// =============================================================================

// Query will wrap `QueryContext` so that we can Scan later
//  Query(cmd.QueryContext(ctx, "..."))
func (brock_sql_box) QueryRow(val *sql.Row, err error) SQLBoxQueryRow {
	return brock_sql_box_query_row{val, err}
}

type SQLBoxQueryRow interface {
	Scan(dest ...any) error
	Err() error
}

type brock_sql_box_query_row struct {
	res *sql.Row
	err error
}

func (x brock_sql_box_query_row) Scan(dest ...any) error {
	return x.res.Scan(dest...)
}

func (x brock_sql_box_query_row) Err() error {
	if x.err == nil {
		x.err = x.res.Err()
	}

	return x.err
}

// =============================================================================

// Transaction will wrap `Begin` so that we can Wrap later
//  Transaction(db.BeginTx(ctx, ...))
//
// Wrap the transaction and ends it with either COMMIT or ROLLBACK
func (brock_sql_box) Transaction(val *sql.Tx, err error) SQLBoxTransaction {
	return &brock_sql_box_begin_tx{new(sync.Once), val, err}
}

type SQLBoxTransaction interface {
	// Wrap the transaction and ends it with either COMMIT or ROLLBACK
	Wrap(tx func() error) error
}

type brock_sql_box_begin_tx struct {
	once *sync.Once
	res  *sql.Tx
	err  error
}

func (x brock_sql_box_begin_tx) Wrap(tx func() error) error {
	if x.err != nil {
		return x.err
	}

	new(sql.Row).Scan()

	fn := Yield(error(ErrInvalidTransaction))
	x.once.Do(func() {
		if err := tx(); err != nil {
			fn = Yield(x.res.Rollback())
			if fn() == nil {
				fn = Yield(err)
			}
		} else {
			fn = Yield(x.res.Commit())
		}
		x.err = fn()
	})
	return fn()
}

// =============================================================================

func (brock_sql) RoundRobin(conns ...SQLConn) SQLConn {
	conns_ := make([]SQLConn, 0)
	for _, v := range conns {
		if v != nil {
			conns_ = append(conns_, v)
		}
	}

	return &brock_sql_roundrobin{conns_, 0, new(sync.Mutex)}
}

type brock_sql_roundrobin struct {
	conns []SQLConn
	index int
	mutex *sync.Mutex
}

func (x *brock_sql_roundrobin) conn(i int) (SQLConn, error) {
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
		return nil, &SQLRoundRobinError{l, i}
	}

	return x.conns[x.index], nil
}

func (x *brock_sql_roundrobin) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	conn, err := x.conn(0)
	if err != nil {
		return nil, err
	}

	return conn.BeginTx(ctx, opts)
}

// Close all databases.
func (x *brock_sql_roundrobin) Close() error {
	errs := make([]error, 0)
	for i := range x.conns {
		errs = append(errs, x.conns[i].Close())
	}

	return Errors(errs...)
}

// PingContext all databases.
func (x *brock_sql_roundrobin) PingContext(ctx context.Context) error {
	errs := make([]error, 0)
	for i := range x.conns {
		errs = append(errs, x.conns[i].PingContext(ctx))
	}

	return Errors(errs...)
}

func (x *brock_sql_roundrobin) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, ErrUnimplemented
}
func (x *brock_sql_roundrobin) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, ErrUnimplemented
}
func (x *brock_sql_roundrobin) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, ErrUnimplemented
}
func (x *brock_sql_roundrobin) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return new(sql.Row)
}

type brock_sql_helper struct{}

const (
	_SELECT   = "SELECT"
	_INSERT   = "INSERT"
	_UPDATE   = "UPDATE"
	_DELETE   = "DELETE"
	_CREATE   = "CREATE"
	_ALTER    = "ALTER"
	_DROP     = "DROP"
	_USE      = "USE"
	_ADD      = "ADD"
	_EXEC     = "EXEC"
	_TRUNCATE = "TRUNCATE"
)

// RemoveComment from sql command.
func (brock_sql_helper) RemoveComment(query string) string {
	commentStartIdx, replaces := -1, []string{}

	for i := range query {
		// we found sql comment
		if i-1 >= 0 && i+1 < len(query) && query[i] == '-' && query[i+1] == '-' && query[i-1] != '-' {
			commentStartIdx = i

			continue
		}

		if commentStartIdx > -1 && query[i] == '\n' {
			replaces = append(replaces, query[commentStartIdx:i])
		}
	}

	for _, v := range replaces {
		query = strings.Replace(query, v, "", 1)
	}

	return strings.TrimSpace(query)
}

// IsMultipleCommand is a naive implementation of checking multiple sql command.
func (x brock_sql_helper) IsMultipleCommand(query string) bool {
	validCount := 0

	for _, query := range strings.Split(query, ";") {
		query = strings.ToUpper(strings.TrimSpace(x.RemoveComment(query)))
		if x.IsValidCommand(query) {
			validCount++
		}
	}

	return validCount > 1
}

// IsSELECTCommand only valid if starts with SELECT.
func (x brock_sql_helper) IsSELECTCommand(query string) bool {
	var ok bool

	query = strings.ToUpper(strings.TrimSpace(x.RemoveComment(query)))
	for _, s := range []string{_SELECT} {
		ok = ok || strings.HasPrefix(query, s) || strings.Contains(query, s)
	}

	return ok
}

// IsDMLCommand only valid if starts with INSERT, UPDATE, DELETE.
func (x brock_sql_helper) IsDMLCommand(query string) bool {
	var ok bool

	query = strings.ToUpper(strings.TrimSpace(x.RemoveComment(query)))
	for _, s := range []string{_INSERT, _UPDATE, _DELETE} {
		ok = ok || strings.HasPrefix(query, s)
	}

	return ok
}

// IsDDLCommand only valid if starts with CREATE, ALTER, DROP, USE, ADD, EXEC, TRUNCATE.
func (x brock_sql_helper) IsDDLCommand(query string) bool {
	var ok bool

	query = strings.ToUpper(strings.TrimSpace(x.RemoveComment(query)))
	for _, s := range []string{_CREATE, _ALTER, _DROP, _USE, _ADD, _EXEC, _TRUNCATE} {
		ok = ok || strings.HasPrefix(query, s)
	}

	return ok
}

func (x brock_sql_helper) IsValidCommand(query string) bool {
	return x.IsSELECTCommand(query) || x.IsDMLCommand(query) || x.IsDDLCommand(query)
}
