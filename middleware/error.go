package middleware

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"tu-xun/common"
	"tu-xun/controller"

	vl "tu-xun/service/validator"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func Error(c *gin.Context) {
	c.Next()
	if len(c.Errors) != 0 {
		err := c.Errors.Last().Err
		switch err := err.(type) {
		case validator.ValidationErrors:
			var errs string
			for _, v := range err.Translate(vl.Trans) {
				errs = fmt.Sprintf("%v,%v", errs, v)
			}
			errorHandle(c, strings.Replace(errs, ",", "", 1))
		case *strconv.NumError, *json.UnmarshalTypeError, *time.ParseError, *xml.SyntaxError:
			errorHandle(c, errors.New("错误或非法的传入参数"))
		default:
			errorHandle(c, err)
		}
	}
}

func errorHandle(c *gin.Context, err any) {
	errType := c.Errors.Last().Type
	errMsg := fmt.Sprintf("%v: %v", common.ErrorMapper[uint64(errType)], err)
	httpStatus := common.HTTPStatus(errType)
	c.JSON(httpStatus, controller.Response{
		Success: false,
		Data:    nil,
		Message: errMsg,
		Code:    uint64(errType),
	})
}
