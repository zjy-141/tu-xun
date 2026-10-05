package validator

import (
	ut "github.com/go-playground/universal-translator"
)

// timingTransZh 为 timing 校验器提供中文错误文案。
// 与 timing 一同处于未启用状态；启用时需在 init.go 的注册表中登记。
func timingTransZh(ut ut.Translator) error {
	return ut.Add("timing", "{0}输入的时间不符合要求", true)
}
