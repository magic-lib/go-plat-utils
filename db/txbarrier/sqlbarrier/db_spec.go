package sqlbarrier

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// DBTypeMysql represents the mysql db type.
	DBTypeMysql = "mysql"
	// DBTypeSqlite represents the sqlite db type.
	DBTypeSqlite = "sqlite"
	// DBTypePostgres represents the postgres db type.
	DBTypePostgres = "postgres"
)

func init() {
	RegisterDBSpecial(DBTypeMysql, &mysqlSpec{})
	RegisterDBSpecial(DBTypeSqlite, &sqliteSpec{})
	RegisterDBSpecial(DBTypePostgres, &postgresSpec{})
}

// DBSpecial defines the dialect acquisition interface of different db types,
// which is convenient to expand to different sql databases.
type DBSpecial interface {
	// GetInsertIgnoreSQL returns the sql dialect
	// that ignores unique key conflicts when inserting data.
	GetInsertIgnoreSQL(table string, fields []string, constraint string) string
}

var dbSpecials = map[string]DBSpecial{}

// RegisterDBSpecial register a DBSpecial into sqlbarrier.
func RegisterDBSpecial(dbType string, special DBSpecial) {
	dbSpecials[dbType] = special
}

type mysqlSpec struct {
}

// GetInsertIgnoreSQL implements the DBSpecial.GetInsertIgnoreSQL for mysql.
func (m *mysqlSpec) GetInsertIgnoreSQL(table string, fields []string, _ string) string {
	valPlaceHold := strings.Repeat("?,", len(fields)-1) + "?"
	return fmt.Sprintf("insert ignore into %s(%s) values(%s)",
		table, strings.Join(fields, ","), valPlaceHold)
}

type sqliteSpec struct {
}

// GetInsertIgnoreSQL implements the DBSpecial.GetInsertIgnoreSQL for sqlite.
func (s *sqliteSpec) GetInsertIgnoreSQL(table string, fields []string, _ string) string {
	valPlaceHold := strings.Repeat("?,", len(fields)-1) + "?"
	return fmt.Sprintf("insert or ignore into %s(%s) values(%s)",
		table, strings.Join(fields, ","), valPlaceHold)
}

type postgresSpec struct {
}

// GetInsertIgnoreSQL implements the DBSpecial.GetInsertIgnoreSQL for postgres.
func (p *postgresSpec) GetInsertIgnoreSQL(table string, fields []string, constraint string) string {
	var valPlaceHold = make([]string, len(fields))
	for i := range fields {
		valPlaceHold[i] = "$" + strconv.Itoa(i+1)
	}
	return fmt.Sprintf("insert into %s(%s) values(%s) on conflict ON CONSTRAINT %s do nothing",
		table, strings.Join(fields, ","), strings.Join(valPlaceHold, ","), constraint)
}
