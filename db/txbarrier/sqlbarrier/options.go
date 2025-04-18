package sqlbarrier

// Option is the Optional configuration for sqlbarrier.
type Option func(*Conn)

// WithDBType sets db type of the driver. Default is DBTypeMysql.
func WithDBType(typ string) Option {
	return func(c *Conn) {
		c.dbType = typ
	}
}

// WithTableName sets the table name for sqlbarrier. Allows using "." to
// separate database name and table name, such as "tdxa.txbarrier".
// Default is "tdxa.txbarrier"
func WithTableName(table string) Option {
	return func(c *Conn) {
		c.table = table
	}
}

// WithUniqConstraint sets the unique constraint name for sqlbarrier.
// The constraint name can't be empty if the db type is DBTypePostgres.
// Default is "barrier_unique_key"
func WithUniqConstraint(uniq string) Option {
	return func(c *Conn) {
		c.constraint = uniq
	}
}
