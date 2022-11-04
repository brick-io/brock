package sdksql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"

	"github.com/lib/pq"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type BeginTx interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (tx *sql.Tx, err error)
}
type ExecContext interface {
	ExecContext(ctx context.Context, query string, args ...any) (res sql.Result, err error)
}
type PingContext interface {
	PingContext(ctx context.Context) (err error)
}
type PrepareContext interface {
	PrepareContext(ctx context.Context, query string) (stmt *sql.Stmt, err error)
}
type QueryContext interface {
	QueryContext(ctx context.Context, query string, args ...any) (rows *sql.Rows, err error)
}
type QueryRowContext interface {
	QueryRowContext(ctx context.Context, query string, args ...any) (row *sql.Row)
}

// Conn is a common interface of *sql.DB and *sql.Conn.
type Conn interface {
	BeginTx
	io.Closer
	PingContext
	TxConn
}

// TxConn is a common interface of *sql.DB, *sql.Conn, and *sql.Tx.
type TxConn interface {
	ExecContext
	PrepareContext
	QueryContext
	QueryRowContext
}

func Open(dsn string) (*sql.DB, error) {
	driverName := strings.Split(dsn+"://", "://")[0]
	switch driverName {
	case "postgres", "postgresql":
		return sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn))), nil
	default:
		return sql.Open(driverName, dsn)
	}
}

func Discard() any { return new([]byte) }

func ArrayPostgreSQL(array any) interface {
	driver.Valuer
	sql.Scanner
} {
	return struct {
		driver.Valuer
		sql.Scanner
	}{
		driver.Valuer(pq.Array(array)),
		sql.Scanner(pgdialect.Array(array)),
	}
}
