package sensitive

import (
	"log"
	"sync"

	"github.com/kirklin/go-swd"
)

var (
	detector *swd.SWD
	once     sync.Once
)

// Init 初始化敏感词检测器（单例，仅执行一次）
func Init() {
	once.Do(func() {
		d, err := swd.New()
		if err != nil {
			log.Fatalf("初始化敏感词检测失败: %v", err)
		}
		detector = d
	})
}

// Detect 检测文本是否包含敏感词，返回 true 表示包含敏感词
func Detect(text string) bool {
	if detector == nil {
		return false
	}
	return detector.Detect(text)
}
