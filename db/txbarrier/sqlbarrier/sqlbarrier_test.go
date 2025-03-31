package sqlbarrier

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"

	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"github.com/magic-lib/go-plat-utils/db/txbarrier"
)

const (
	sqliteCreateTxbarrier = `
create table txbarrier (
  id integer PRIMARY KEY AUTOINCREMENT,
  xid varchar(128),
  branch_id varchar(128),
  op varchar(45),
  reason varchar(45),
  UNIQUE(xid, branch_id, op)
)`
	sqliteCreateTestTable = `
create table test_table (
  id integer PRIMARY KEY AUTOINCREMENT,
  identify varchar(128),
  content varchar(128)
)
`
	testBranch     = "branch_test"
	testDriverName = "sqlite3-test-barrier"
)

func TestSQLBarrier(t *testing.T) {
	registerTestDriver()
	t.Run("test: idempotent", func(t *testing.T) {
		db, err := newTestDB("idempotent_test")
		require.Nil(t, err)

		barrierTest(t, db, "1", txbarrier.Try, "idempotent")
		barrierTest(t, db, "2", txbarrier.Confirm, "idempotent")
		barrierTest(t, db, "3", txbarrier.Cancel, "idempotent")
	})
	t.Run("test: empty compensation and suspension", func(t *testing.T) {
		db, err := newTestDB("empty_compensation_and_suspension")
		require.Nil(t, err)

		emptyCompensationAndSuspensionTest(t, db)
	})
	t.Run("test: normal", func(t *testing.T) {
		db, err := newTestDB("normal")
		require.Nil(t, err)

		barrierTest(t, db, "5", txbarrier.Confirm, "normal")
		barrierTest(t, db, "6", txbarrier.Cancel, "normal")
	})
	t.Run("test: context without Barrier", func(t *testing.T) {
		db, err := newTestDB("no_barrier")
		require.Nil(t, err)

		noBarrierInfoTest(t, db)
	})
	t.Run("test: business error", func(t *testing.T) {
		db, err := newTestDB("business_err")
		require.Nil(t, err)

		businessErrorTest(t, db, "7")
	})
}

func registerTestDriver() {
	opts := []Option{
		WithDBType(DBTypeSqlite),
		WithTableName("txbarrier"),
		WithUniqConstraint("xid_branch_id_op"),
	}
	drv := NewDriver(&sqlite3.SQLiteDriver{}, opts...)
	sql.Register(testDriverName, drv)
}
func newTestDB(name string) (*sql.DB, error) {
	db, err := sql.Open(testDriverName, fmt.Sprintf("file:%s?mode=memory&cache=shared", name))
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(sqliteCreateTxbarrier)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(sqliteCreateTestTable)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func barrierTest(t *testing.T, db *sql.DB, xid string, op txbarrier.Operation, mode string) {
	// try
	ctx := txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
		XID:      xid,
		BranchID: testBranch,
		TransTyp: "tcc",
		Op:       txbarrier.Try,
	})
	tx, err := db.BeginTx(ctx, nil)
	require.Nil(t, err)
	ret, err := tx.ExecContext(ctx, "insert or ignore into test_table(identify, content) values(?,?)", xid, txbarrier.Try)
	require.Nil(t, err)
	affected, err := ret.RowsAffected()
	require.Nil(t, err)
	require.NotZero(t, affected)
	require.Nil(t, tx.Commit())

	if op != txbarrier.Try {
		// request arrives
		ctx = txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
			XID:      xid,
			BranchID: testBranch,
			TransTyp: "tcc",
			Op:       op,
		})
		tx, err = db.BeginTx(ctx, nil)
		require.Nil(t, err)
		ret, err = tx.ExecContext(ctx, "update test_table set content = ? where identify=?", op, xid)
		require.Nil(t, err)
		affected, err = ret.RowsAffected()
		require.Nil(t, err)
		require.NotZero(t, affected)
		require.Nil(t, tx.Commit())
	}

	if mode == "idempotent" {
		// repeats the request
		tx, err = db.BeginTx(ctx, nil)
		require.Equal(t, txbarrier.ErrDuplicationOrSuspension, err)
		require.Nil(t, tx)
	}

	if mode == "normal" {
		rows, err := db.QueryContext(ctx, "select content from test_table where identify=?", xid)
		require.Nil(t, err)
		defer rows.Close()
		require.True(t, rows.Next())
		content := ""
		require.Nil(t, rows.Scan(&content))
		require.Equal(t, string(op), content)
	}
}

func emptyCompensationAndSuspensionTest(t *testing.T, db *sql.DB) {
	// Cancel request arrives before Try.
	ctx := txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
		XID:      "4",
		BranchID: testBranch,
		TransTyp: "tcc",
		Op:       txbarrier.Cancel,
	})
	tx, err := db.BeginTx(ctx, nil)
	require.Equal(t, txbarrier.ErrEmptyCompensation, err)
	require.Nil(t, tx)

	// Try request arrives after Cancel.
	ctx = txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
		XID:      "4",
		BranchID: testBranch,
		TransTyp: "tcc",
		Op:       txbarrier.Try,
	})
	tx, err = db.BeginTx(ctx, nil)
	require.Equal(t, txbarrier.ErrDuplicationOrSuspension, err)
	require.Nil(t, tx)
}

func noBarrierInfoTest(t *testing.T, db *sql.DB) {
	ctx := context.TODO()
	tx, err := db.BeginTx(ctx, nil)
	require.Nil(t, err)
	ret, err := tx.ExecContext(ctx, "insert or ignore into test_table(identify, content) values(?,?)", "1", "2")
	require.Nil(t, err)
	affected, err := ret.RowsAffected()
	require.Nil(t, err)
	require.NotZero(t, affected)
	require.Nil(t, tx.Commit())

	rows, err := db.QueryContext(ctx, "select content from test_table where identify=?", "1")
	require.Nil(t, err)
	defer rows.Close()
	require.True(t, rows.Next())
	content := ""
	require.Nil(t, rows.Scan(&content))
	require.Equal(t, "2", content)
}

func businessErrorTest(t *testing.T, db *sql.DB, xid string) {
	ctx := txbarrier.NewCtxWithBarrier(context.TODO(), &txbarrier.Barrier{
		XID:      xid,
		BranchID: testBranch,
		TransTyp: "tcc",
		Op:       txbarrier.Try,
	})
	tx, err := db.BeginTx(ctx, nil)
	require.Nil(t, err)
	ret, err := tx.ExecContext(ctx, "insert or ignore into test_table(identify, content) values(?,?)", xid, txbarrier.Try)
	require.Nil(t, err)
	affected, err := ret.RowsAffected()
	require.Nil(t, err)
	require.NotZero(t, affected)

	// assume occurred business error, rollback
	require.Nil(t, tx.Rollback())

	rows, err := db.QueryContext(ctx, "select content from test_table where identify=?", xid)
	require.Nil(t, err)
	defer rows.Close()
	require.False(t, rows.Next())
}

const (
	testNoConnBeginTx                 = "noConnBeginTx:123@tcp(localhost:3306)/test_db"
	testExecerQueryCtxSessionResetter = "allImpl:123@tcp(localhost:3306)/test_db"
	testExecerQueryCtx                = "excecerQueryCtx:123@tcp(localhost:3306)/test_db"
	testExecer                        = "execer:123@tcp(localhost:3306)/test_db"
	testQueryCtx                      = "queryCtx:123@tcp(localhost:3306)/test_db"
)

type mockDriver struct {
}

func (m *mockDriver) Open(name string) (driver.Conn, error) {
	switch name {
	case testNoConnBeginTx:
		return struct{ driver.Conn }{}, nil
	case testExecerQueryCtxSessionResetter:
		return struct {
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
		}{}, nil
	case testExecerQueryCtx:
		return struct {
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.QueryerContext
		}{}, nil
	case testExecer:
		return struct {
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
		}{}, nil
	case testQueryCtx:
		return struct {
			driver.Conn
			driver.ConnBeginTx
			driver.QueryerContext
		}{}, nil
	}

	return nil, fmt.Errorf("test error")
}

func TestDriver_Open(t *testing.T) {
	mdrv := &mockDriver{}
	drv := NewDriver(mdrv)

	t.Run("normal error", func(t *testing.T) {
		c, err := drv.Open("test error")
		require.Error(t, err)
		require.Nil(t, c)

	})
	t.Run("no ConnBeginTx", func(t *testing.T) {
		c, err := drv.Open(testNoConnBeginTx)
		require.Error(t, err)
		require.Nil(t, c)
	})
	t.Run("ExecerQueryCtxSessionResetter", func(t *testing.T) {
		c, err := drv.Open(testExecerQueryCtxSessionResetter)
		require.NoError(t, err)
		_, ok := c.(*ExecerQueryerContextWithSR)
		require.True(t, ok)
	})
	t.Run("ExecerQueryCtx", func(t *testing.T) {
		c, err := drv.Open(testExecerQueryCtx)
		require.NoError(t, err)
		_, ok := c.(*ExecerQueryerContext)
		require.True(t, ok)
	})
	t.Run("Execer", func(t *testing.T) {
		c, err := drv.Open(testExecer)
		require.NoError(t, err)
		_, ok := c.(*ExecerContext)
		require.True(t, ok)
	})
	t.Run("queryerContext", func(t *testing.T) {
		c, err := drv.Open(testQueryCtx)
		require.NoError(t, err)
		_, ok := c.(*QueryerContext)
		require.True(t, ok)
	})
}

var (
	errTestWithoutContexter = errors.New("without contexter")
	errTestWithContexter    = errors.New("with contexter")
)

type mockConn struct {
	driver.Conn
}

func (m *mockConn) Prepare(_ string) (driver.Stmt, error) {
	return nil, errTestWithoutContexter
}

func (m *mockConn) Exec(_ string, _ []driver.Value) (driver.Result, error) {
	return nil, errTestWithoutContexter
}

func (m *mockConn) Query(_ string, _ []driver.Value) (driver.Rows, error) {
	return nil, errTestWithoutContexter
}

type mockConnWithContexter struct {
	driver.Conn
}

func (m *mockConnWithContexter) PrepareContext(ctx context.Context, q string) (driver.Stmt, error) {
	return nil, errTestWithContexter
}

func (m *mockConnWithContexter) ExecContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
	return nil, errTestWithContexter
}

func (m *mockConnWithContexter) QueryContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
	return nil, errTestWithContexter
}

func TestConn_PrepareContext(t *testing.T) {
	t.Run("without ConnPrepareContext", func(t *testing.T) {
		conn := &Conn{Conn: &mockConn{}}
		_, err := conn.PrepareContext(context.TODO(), "--")
		require.Equal(t, errTestWithoutContexter, err)
	})
	t.Run("with ConnPrepareContext", func(t *testing.T) {
		conn := &Conn{Conn: &mockConnWithContexter{}}
		_, err := conn.PrepareContext(context.TODO(), "--")
		require.Equal(t, errTestWithContexter, err)
	})
	t.Run("without ConnPrepareContext and context canceled", func(t *testing.T) {
		conn := &Conn{Conn: &mockConn{}}
		ctx, cancel := context.WithCancel(context.TODO())
		cancel()
		_, err := conn.PrepareContext(ctx, "--")
		require.Equal(t, context.Canceled, err)
	})
}

func TestExecerContext_ExecContext(t *testing.T) {
	named := []driver.NamedValue{
		{Ordinal: 1, Value: "test"},
	}
	t.Run("without ExecerContext", func(t *testing.T) {
		exec := &ExecerContext{Conn: &Conn{Conn: &mockConn{}}}
		_, err := exec.ExecContext(context.TODO(), "--", named)
		require.Equal(t, errTestWithoutContexter, err)
	})
	t.Run("with ExecerContext", func(t *testing.T) {
		exec := &ExecerContext{Conn: &Conn{Conn: &mockConnWithContexter{}}}
		_, err := exec.ExecContext(context.TODO(), "--", named)
		require.Equal(t, errTestWithContexter, err)
	})
	t.Run("without ExecerContext and context canceled", func(t *testing.T) {
		exec := &ExecerContext{Conn: &Conn{Conn: &mockConn{}}}
		ctx, cancel := context.WithCancel(context.TODO())
		cancel()
		_, err := exec.ExecContext(ctx, "--", named)
		require.Equal(t, context.Canceled, err)
	})
}

func TestQueryerContext_QueryContext(t *testing.T) {
	named := []driver.NamedValue{
		{Ordinal: 1, Value: "test"},
	}
	t.Run("without QueryerContext", func(t *testing.T) {
		q := &QueryerContext{Conn: &Conn{Conn: &mockConn{}}}
		_, err := q.QueryContext(context.TODO(), "--", named)
		require.Equal(t, errTestWithoutContexter, err)
	})
	t.Run("with QueryerContext", func(t *testing.T) {
		q := &QueryerContext{Conn: &Conn{Conn: &mockConnWithContexter{}}}
		_, err := q.QueryContext(context.TODO(), "--", named)
		require.Equal(t, errTestWithContexter, err)
	})
	t.Run("without QueryerContext and context canceled", func(t *testing.T) {
		q := &QueryerContext{Conn: &Conn{Conn: &mockConn{}}}
		ctx, cancel := context.WithCancel(context.TODO())
		cancel()
		_, err := q.QueryContext(ctx, "--", named)
		require.Equal(t, context.Canceled, err)
	})
}

func Test_namedValueToValue(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		named := []driver.NamedValue{
			{Ordinal: 1, Value: "test"},
		}
		args, err := namedValueToValue(named)
		require.NoError(t, err)
		require.Len(t, args, 1)
	})
	t.Run("failed", func(t *testing.T) {
		named := []driver.NamedValue{
			{Name: "name", Ordinal: 1, Value: "test"},
		}
		args, err := namedValueToValue(named)
		require.Error(t, err)
		require.Len(t, args, 0)
	})
}
