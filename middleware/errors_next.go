package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin"
)

// ErrorsNext
// igin.AddStatusError(context, igin.NewHTTPError(201, "test error"))
// 预期:status:201;error:code=201, message=test error"
//
// igin.AddStatusError(context, fmt.Errorf("test error"))
// 预期:status:500;error:test error
//
// igin.AddStatusError(context, fmt.Errorf("test error"), 201)
// 预期:status:201;error:test error
func ErrorsNext(errorHandlers ...ErrorHandler) gin.HandlerFunc {
	var errorHandler ErrorHandler
	if len(errorHandlers) > 0 {
		errorHandler = errorHandlers[0]
	} else {
		errorHandler = func(c *gin.Context, err error, statusCodes ...int) {
			statusCode := http.StatusInternalServerError
			if len(statusCodes) > 0 {
				statusCode = statusCodes[0]
			}
			json := &igin.JsonResponse{Code: statusCode}
			json.SetErr(err)
			igin.ContextIResponse(c, json, func(c *gin.Context, response igin.IResponse) int {
				var finalStatus int
				//即将响应到状态码
				// 参数验证失败响应到状态码是400
				status := c.Writer.Status()
				//response中定义到状态码
				respStatus := response.StatusCode()
				// 首先判断即将响应状态码和response中到状态码是否一致
				//不一致情况
				//即将响应状态码如果是200 就返回response中状态码
				//否则就返回即将响应到状态码
				if respStatus != status && status == http.StatusOK {
					finalStatus = respStatus
				} else {
					finalStatus = status
				}
				return finalStatus
			})
		}
	}
	return func(c *gin.Context) {
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				errorHandler(c, err.Err)
				return
			}
		}
		c.Next()
	}
}
