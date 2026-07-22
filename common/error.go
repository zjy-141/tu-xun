package common

import "github.com/gin-gonic/gin"

const (
	ParamErr    gin.ErrorType = iota + 3
	SysErr
	OpErr
	AuthErr
	LevelErr
	ConflictErr
)

var ErrorMapper = map[uint64]string{
	1: "内部错误",
	2: "公开错误",
	3: "参数错误",
	4: "系统错误",
	5: "操作错误",
	6: "鉴权错误",
	7: "权限错误",
	8: "冲突错误",
}

// HTTPStatus 根据错误类型返回对应的 HTTP 状态码
func HTTPStatus(errType gin.ErrorType) int {
	switch errType {
	case AuthErr:
		return 401
	case LevelErr:
		return 403
	case ConflictErr:
		return 409
	case ParamErr:
		return 400
	case OpErr:
		return 404
	default:
		return 200
	}
}

// ErrNew 创建带错误类型的 gin.Error，用于中间件统一处理
func ErrNew(err error, errType gin.ErrorType) error {
	err = &gin.Error{
		Err:  err,
		Type: errType,
	}
	return err
}
