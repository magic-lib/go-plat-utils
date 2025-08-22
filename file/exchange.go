package file

type ExchangeType string

const (
	ExchangeTypeTEXT ExchangeType = "text" //文本文件
)

// Exchanger 从文本，csv等批量导入数据
type Exchanger interface {
	ExchangeList(exchangeType ExchangeType) ([][]string, error)
	ExchangeOne(exchangeType ExchangeType) (string, error)
}
