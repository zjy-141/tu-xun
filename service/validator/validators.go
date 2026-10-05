package validator

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// timing 时间校验器（示例/扩展位）。
// 注意：当前全库没有任何 struct tag 使用它，且未在 validatorHandleRouter 注册，
// 因此不会生效。真正撑住入参校验的是各处 binding:"..." tag（约 137 处）。
// 保留函数体作为自定义校验规则的实现范例；如需启用，请在 init.go 的注册表中登记。
func timing(fl validator.FieldLevel) bool {
	if date, ok := fl.Field().Interface().(time.Time); ok {
		today := time.Now()
		if today.After(date) {
			return false
		}
	}
	return true
}
