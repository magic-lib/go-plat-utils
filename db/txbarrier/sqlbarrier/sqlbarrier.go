// Package sqlbarrier is designed to solve the timing problem
// of accessing RM(Resource Manager) based on SQL databases
// in distributed transactions, while avoiding excessive
// intrusion into business code.
// To hit the goals, sqlbarrier implements the database/sql/driver.Driver,
// and blocks the duplicated request, empty compensation or hanging
// request in the Driver.
package sqlbarrier

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/magic-lib/go-plat-utils/db/txbarrier"
)

var (
	defaultConn = Conn{dbType: DBTypeMysql, table: "tdxa.txbarrier", constraint: "barrier_unique_key"}
)

// Driver implements the database/sql/driver.Driver.
type Driver struct {
	driver.Driver
	opts []Option
}

// NewDriver creates a new driver.Driver, it takes a vendor specific driver.
func NewDriver(drv driver.Driver, opts ...Option) driver.Driver {
	return &Driver{Driver: drv, opts: opts}
}

// Open opens a connection.
func (d *Driver) Open(dsn string) (driver.Conn, error) {
	conn, err := d.Driver.Open(dsn)
	if err != nil {
		return conn, err
	}

	// Conn that doesn't implement driver.ConnBeginTx are not supported.
	if !isConnBeginTx(conn) {
		return nil, fmt.Errorf("driver must implements driver.ConnBeginTx")
	}

	wrapped := &Conn{Conn: conn, dbType: DBTypeMysql, table: defaultConn.table, constraint: defaultConn.constraint}
	for _, o := range d.opts {
		o(wrapped)
	}

	// Returns an instance according wrapped underlying driver implementations.
	if isExecer(conn) && isQueryContext(conn) && isSessionResetter(conn) {
		return &ExecerQueryerContextWithSR{
			Conn: wrapped,
			ExecerQueryerContext: &ExecerQueryerContext{
				Conn:           wrapped,
				ExecerContext:  &ExecerContext{wrapped},
				QueryerContext: &QueryerContext{wrapped},
			},
			SessionResetter: &SessionResetter{wrapped},
		}, nil
	} else if isExecer(conn) && isQueryContext(conn) {
		return &ExecerQueryerContext{
			Conn:           wrapped,
			ExecerContext:  &ExecerContext{wrapped},
			QueryerContext: &QueryerContext{wrapped},
		}, nil

	} else if isExecer(conn) {
		return &ExecerContext{wrapped}, nil
	} else if isQueryContext(conn) {
		return &QueryerContext{wrapped}, nil
	}

	return wrapped, nil
}

// Conn implements the driver.Conn
type Conn struct {
	driver.Conn
	dbType     string
	table      string
	constraint string
}

// PrepareContext implements the driver.ConnPrepareContext
func (c *Conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if pc, ok := c.Conn.(driver.ConnPrepareContext); ok {
		return pc.PrepareContext(ctx, query)
	}

	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return c.Conn.Prepare(query)
}

// BeginTx implements the driver.ConnBeginTx.
// The core of sqlbarrier is mainly implemented in this method.
func (c *Conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	tx, err := c.Conn.(driver.ConnBeginTx).BeginTx(ctx, opts)
	if err != nil {
		return tx, err
	}

	err = c.barrierCheck(ctx)
	if errors.Is(err, txbarrier.ErrEmptyCompensation) {
		_ = tx.Commit()
		return nil, err
	}
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	return tx, nil
}

// barrierCheck returns txbarrier.ErrDuplicationOrSuspension or
// txbarrier.ErrEmptyCompensation if occurs duplicated request,
// empty compensation or hanging request.
// When the caller encounters txbarrier.ErrEmptyCompensation,
// tx.Commit should be called; for other errors, tx.Rollback
// should be called.
func (c *Conn) barrierCheck(ctx context.Context) error {
	b := txbarrier.BarrierFromCtx(ctx)
	if !b.Valid() {
		return nil
	}

	affected, err := c.insertDB(ctx, b.XID, b.BranchID, b.Op, string(b.Op))
	if err != nil {
		return err
	}
	if affected == 0 {
		return txbarrier.ErrDuplicationOrSuspension
	}

	if b.Op == txbarrier.Cancel {
		affected, err = c.insertDB(ctx, b.XID, b.BranchID, txbarrier.Try, string(b.Op))
		if err != nil {
			return err
		}
		// The insertion of txbarrier.Try is successful, indicating that txbarrier.Cancel
		// arrived earlier than txbarrier.Try.
		if affected > 0 {
			return txbarrier.ErrEmptyCompensation
		}
	}

	return nil
}

func (c *Conn) insertDB(ctx context.Context,
	xid, branchID string, op txbarrier.Operation, reason string) (int64, error) {
	sqlStr, err := c.getInsertSQL()
	if err != nil {
		return 0, err
	}

	stmt, err := c.PrepareContext(ctx, sqlStr)
	if err != nil {
		return 0, err
	}

	var result driver.Result
	var args = []driver.NamedValue{
		{Ordinal: 1, Value: xid},
		{Ordinal: 2, Value: branchID},
		{Ordinal: 3, Value: string(op)},
		{Ordinal: 4, Value: reason},
	}
	if stmtCtx, ok := stmt.(driver.StmtExecContext); ok {
		result, err = stmtCtx.ExecContext(ctx, args)
		if err != nil {
			return 0, err
		}
	} else {
		a, _ := namedValueToValue(args)
		result, err = stmt.Exec(a) // nolint
		if err != nil {
			return 0, err
		}
	}

	return result.RowsAffected()
}

var barrierFields = []string{"xid", "branch_id", "op", "reason"}

func (c *Conn) getInsertSQL() (string, error) {
	dbSepc := dbSpecials[c.dbType]
	if dbSepc == nil {
		return "", fmt.Errorf("sqlbarrier: unregistered db type `%s`", c.dbType)
	}

	return dbSepc.GetInsertIgnoreSQL(c.table, barrierFields, c.constraint), nil
}

// ExecerContext implements the driver.ExecerContext.
type ExecerContext struct {
	*Conn
}

// ExecContext implements the driver.ExecerContext.
func (e *ExecerContext) ExecContext(ctx context.Context,
	query string, args []driver.NamedValue) (driver.Result, error) {
	switch execer := e.Conn.Conn.(type) {
	case driver.ExecerContext:
		return execer.ExecContext(ctx, query, args)
	case driver.Execer: // nolint
		dargs, err := namedValueToValue(args)
		if err != nil {
			return nil, err
		}

		select {
		default:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return execer.Exec(query, dargs)
	default:
		// never happen
		return nil, errors.New("driver.Conn implementations neither ExecerContext nor Execer")
	}
}

// QueryerContext implements driver.QueryerContext
type QueryerContext struct {
	*Conn
}

// QueryContext implements driver.QueryerContext
func (q *QueryerContext) QueryContext(ctx context.Context,
	query string, args []driver.NamedValue) (driver.Rows, error) {
	switch qry := q.Conn.Conn.(type) {
	case driver.QueryerContext:
		return qry.QueryContext(ctx, query, args)
	case driver.Queryer: // nolint
		dargs, err := namedValueToValue(args)
		if err != nil {
			return nil, err
		}

		select {
		default:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return qry.Query(query, dargs)
	default:
		// never happen
		return nil, errors.New("driver.Conn implementations neither QueryerContext nor Queryer")
	}
}

// ExecerQueryerContext implements driver.ExecerContext
// and driver.QueryerContext.
type ExecerQueryerContext struct {
	*Conn
	*ExecerContext
	*QueryerContext
}

// ExecerQueryerContextWithSR implements driver.ExecerContext,
// driver.QueryerContext and driver.SessionResetter.
type ExecerQueryerContextWithSR struct {
	*Conn
	*ExecerQueryerContext
	*SessionResetter
}

// SessionResetter implements driver.SessionResetter.
type SessionResetter struct {
	*Conn
}

// ResetSession implements driver.SessionResetter.
func (s *SessionResetter) ResetSession(ctx context.Context) error {
	c, _ := s.Conn.Conn.(driver.SessionResetter)
	return c.ResetSession(ctx)
}

// namedValueToValue copied from database/sql
func namedValueToValue(named []driver.NamedValue) ([]driver.Value, error) {
	dargs := make([]driver.Value, len(named))
	for n, param := range named {
		if len(param.Name) > 0 {
			return nil, errors.New("sql: driver does not support the use of Named Parameters")
		}
		dargs[n] = param.Value
	}
	return dargs, nil
}
