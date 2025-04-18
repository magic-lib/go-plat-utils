package sqlbarrier

import "database/sql/driver"

func isConnBeginTx(conn driver.Conn) bool {
	_, ok := conn.(driver.ConnBeginTx)
	return ok
}

func isExecer(conn driver.Conn) bool {
	switch conn.(type) {
	case driver.ExecerContext, driver.Execer: // nolint
		return true
	default:
		return false
	}
}

func isQueryContext(conn driver.Conn) bool {
	switch conn.(type) {
	case driver.QueryerContext, driver.Queryer: // nolint
		return true
	default:
		return false
	}
}

func isSessionResetter(conn driver.Conn) bool {
	_, ok := conn.(driver.SessionResetter)
	return ok
}
