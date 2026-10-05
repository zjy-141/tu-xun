package sensitive

import (
	"errors"
	"log"
	"sync"

	"github.com/kirklin/go-swd"
)

// ErrNotInitialized 检测器未初始化。安全组件失效时必须 fail-closed，
// 由调用方按「拒绝」处理，而不是静默放行全部内容。
var ErrNotInitialized = errors.New("敏感词检测器未初始化")

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

// Detect 检测文本是否包含敏感词。返回 (true, nil) 表示命中敏感词。
// 检测器不可用时返回 error，调用方应据此拒绝请求（fail-closed）。
func Detect(text string) (bool, error) {
	if detector == nil {
		return false, ErrNotInitialized
	}
	return detector.Detect(text), nil
}
