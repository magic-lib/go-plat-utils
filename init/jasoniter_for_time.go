package init

import (
	"fmt"
	"github.com/json-iterator/go/extra"
	"github.com/magic-lib/go-plat-utils/conv"
	"time"
	"unsafe"
)

const (
	fullTimeForm = "2006-01-02 15:04:05"
)

// timeCodec 时间格式转换
type timeCodec struct {
}

// Decode 转码
func (codec *timeCodec) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	var ok bool
	s := iter.ReadString()
	*((*time.Time)(ptr)), ok = conv.Time(s)
	if !ok {
		iter.ReportError("decode time.Time", fmt.Sprint(s, " is not valid time format"))
	}
}

// IsEmpty 是否为空时间
func (codec *timeCodec) IsEmpty(ptr unsafe.Pointer) bool {
	ts := *((*time.Time)(ptr))
	return ts.UnixNano() == 0
}

// Encode 转码
func (codec *timeCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	ts := *((*time.Time)(ptr))
	stream.WriteString(ts.Format(fullTimeForm))
}

func initTime() {
	jsoniter.RegisterTypeEncoder("time.Time", &timeCodec{})
	jsoniter.RegisterTypeDecoder("time.Time", &timeCodec{})

	//php兼容模式[]，{}
	extra.RegisterFuzzyDecoders()
}
