// Package conn 连接参数
package conn

// Connect 数据连接对象
type Connect struct {
	Driver   string         `json:"driver,omitempty"`
	Protocol string         `json:"protocol,omitempty"`
	Host     string         `json:"host,omitempty"`
	Port     string         `json:"port,omitempty"`
	Username string         `json:"username,omitempty"`
	Password string         `json:"password,omitempty"`
	Database string         `json:"database,omitempty"`
	Extend   map[string]any `json:"extend,omitempty"`
}

// ConnFunc 数据连接参数
type connOption func(*Connect)

// NewOption 新增
func NewOption() connOption {
	return func(*Connect) {}
}

// DialDriver 连接类型
func (c connOption) DialDriver(driver string) connOption {
	return func(do *Connect) {
		c(do)
		do.Driver = driver
	}
}

// DialProtocol 连接协议
func (c connOption) DialProtocol(protocol string) connOption {
	return func(do *Connect) {
		c(do)
		do.Protocol = protocol
	}
}

// DialHostPort 连接ip和端口号
func (c connOption) DialHostPort(host string, port string) connOption {
	return func(do *Connect) {
		c(do)
		do.Host = host
		do.Port = port
	}
}

// DialDatabase 连接库
func (c connOption) DialDatabase(db string) connOption {
	return func(do *Connect) {
		c(do)
		do.Database = db
	}
}

// DialUserNamePassword 连接用户名和密码
func (c connOption) DialUserNamePassword(username string, password string) connOption {
	return func(do *Connect) {
		c(do)
		do.Username = username
		do.Password = password
	}
}

// DialExtend 扩展函数
func (c connOption) DialExtend(ext map[string]any) connOption {
	return func(do *Connect) {
		c(do)

		if do.Extend == nil {
			do.Extend = make(map[string]any)
		}
		if ext == nil {
			return
		}

		for k, v := range ext {
			do.Extend[k] = v
		}
	}
}
