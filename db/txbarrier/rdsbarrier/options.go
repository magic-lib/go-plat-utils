package rdsbarrier

// Option is the optional configuration for rdsbarrier.
type Option func(*Hook)

// WithTimeout sets the timeout for barrier data. Barrier data that times out will be eliminated.
// Businesses should set the timeout for barrier data according to their needs to prevent transactions
// from not being completed after the data has been eliminated. The default timeout value is 24h.
//
// The rdsbarrier will generate barrier data which is based on xid in redis to records the state of
// distributed transaction branch requests. By using barrier data, rdsbarrier can correctly handle
// duplicated, hanging and empty compensation requests.
//
// It's important to note that setting a too long timeout for barrier data can result in long-term
// occupation of redis memory, while a too short timeout may cause the barrier logic to fail, leading
// to incorrect protection of transaction branch requests.
//
// It is recommended for businesses to set a reasonable timeout for barrier data based on their needs,
// generally slightly longer than the overall timeout of the distributed transaction.
func WithTimeout(seconds int) Option {
	return func(hook *Hook) {
		hook.timeout = seconds
	}
}

// WithClusterMode turns on or off the cluster mode of barrier. In cluster mode,
// rdsbarrier checks the hashtag of the key(s) to be operated on and generates
// the same hashtag for the barrier data, ensuring that the barrier data and the
// business data are located in the same Redis slot.
func WithClusterMode(enable bool) Option {
	return func(hook *Hook) {
		hook.enableCluster = enable
	}
}
