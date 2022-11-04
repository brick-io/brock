package sdksql

import (
	"database/sql"
	"sync"

	"github.com/brick-io/brock/sdk"
)

var (
	ErrInvalidArguments   = sdk.Errorf("brock/sdksql: invalid arguments for scan")
	ErrInvalidTransaction = sdk.Errorf("brock/sdksql: invalid transaction")
	ErrNoColumns          = sdk.Errorf("brock/sdksql: no columns returned")
)

type MismatchColumnsError struct{ Col, Dst int }

func (err *MismatchColumnsError) Error() string {
	return sdk.Sprintf("brock/sdksql: mismatch %d columns on %d destinations",
		err.Col,
		err.Dst,
	)
}

// =============================================================================

//nolint:gochecknoglobals
var Wrap wrap

type wrap struct{}

// Exec will wrap `ExecContext` so that we can Scan later
//
//	Exec(cmd.ExecContext(ctx, "..."))
func (wrap) Exec(val sql.Result, err error) WrapExec {
	return wrapExec{val, err}
}

type WrapExec interface {
	// Scan the result of ExecContext that usually return numbers of rowsAffected
	// and lastInsertID.
	Scan(rowsAffected *int, lastInsertID *int) error
}

type wrapExec struct {
	res sql.Result
	err error
}

func (x wrapExec) Scan(rowsAffected *int, lastInsertID *int) error {
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
//
//	Query(cmd.QueryContext(ctx, "..."))
func (wrap) Query(val *sql.Rows, err error) WrapQuery {
	return wrapQuery{val, err}
}

type WrapQuery interface {
	// Scan accept do, a func that accept `i int` as index and returning list
	// of pointers.
	//  pointers == nil   // break the loop
	//  len(pointers) < 1 // skip the current loop
	//  len(pointers) > 0 // assign the pointer, MUST be same as the length of columns
	Scan(row func(i int) (pointers []any)) error
}

type wrapQuery struct {
	res *sql.Rows
	err error
}

func (x wrapQuery) Scan(row func(i int) []any) error {
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
			return &MismatchColumnsError{len(cols), len(dest)}
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
//
//	Query(cmd.QueryContext(ctx, "..."))
func (wrap) QueryRow(val *sql.Row, err error) WrapQueryRow {
	return wrapQueryRow{val, err}
}

type WrapQueryRow interface {
	Scan(dest ...any) error
	Err() error
}

type wrapQueryRow struct {
	res *sql.Row
	err error
}

func (x wrapQueryRow) Scan(dest ...any) error {
	return x.res.Scan(dest...)
}

func (x wrapQueryRow) Err() error {
	if x.err == nil {
		x.err = x.res.Err()
	}

	return x.err
}

// =============================================================================

// Transaction will wrap `Begin` so that we can Wrap later
//
//	Transaction(db.BeginTx(ctx, ...))
//
// Wrap the transaction and ends it with either COMMIT or ROLLBACK.
func (wrap) Transaction(val *sql.Tx, err error) WrapTransaction {
	return &wrapTransaction{new(sync.Once), val, err}
}

type WrapTransaction interface {
	// Do the transaction and ends it with either COMMIT or ROLLBACK
	Do(tx func() error) error
}

type wrapTransaction struct {
	once *sync.Once
	res  *sql.Tx
	err  error
}

func (x wrapTransaction) Do(tx func() error) error {
	if x.err != nil {
		return x.err
	}

	_ = new(sql.Row).Scan()

	fn := sdk.Yield(error(ErrInvalidTransaction))

	x.once.Do(func() {
		if err := tx(); err != nil {
			fn = sdk.Yield(x.res.Rollback())
			if fn() == nil {
				fn = sdk.Yield(err)
			}
		} else {
			fn = sdk.Yield(x.res.Commit())
		}
		x.err = fn()
	})

	return fn()
}
