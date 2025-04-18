package sqlbarrier

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_mysqlSpec_GetInsertIgnoreSQL(t *testing.T) {
	expected := "insert ignore into tdxa.txbarrier(xid,branch_id,op,reason) values(?,?,?,?)"
	spec := &mysqlSpec{}
	sqlStr := spec.GetInsertIgnoreSQL(defaultConn.table, barrierFields, defaultConn.constraint)
	require.Equal(t, expected, sqlStr)
}

func Test_sqliteSpec_GetInsertIgnoreSQL(t *testing.T) {
	expected := "insert or ignore into tdxa.txbarrier(xid,branch_id,op,reason) values(?,?,?,?)"
	spec := &sqliteSpec{}
	sqlStr := spec.GetInsertIgnoreSQL(defaultConn.table, barrierFields, defaultConn.constraint)
	require.Equal(t, expected, sqlStr)
}

func Test_postgresSpec_GetInsertIgnoreSQL(t *testing.T) {
	expected := "insert into tdxa.txbarrier(xid,branch_id,op,reason) values($1,$2,$3,$4) on conflict ON CONSTRAINT barrier_unique_key do nothing" // nolint
	spec := &postgresSpec{}
	sqlStr := spec.GetInsertIgnoreSQL(defaultConn.table, barrierFields, defaultConn.constraint)
	require.Equal(t, expected, sqlStr)
}
