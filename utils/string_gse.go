package utils

import (
	"github.com/go-ego/gse" //分词
	"sync"
)

var (
	once sync.Once
	seg  gse.Segmenter
)

func initSegmenter() {
	once.Do(func() {
		_ = seg.LoadDict()
		_ = seg.LoadDict("zh_s")
		_ = seg.LoadDictEmbed("zh_s")
	})
}
