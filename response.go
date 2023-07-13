package igin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/pkg6/igin/xerror"
	"net/http"
)

type (
	// IResponse 响应状态码根据你传入的状态码走
	IResponse interface {
		//WithStatusCode 携带状态码
		WithStatusCode(statusCode int)
		//StatusCode 响应状态码
		StatusCode() int
		//Abort 是否抛出异常
		Abort() bool
		//Render 渲染数据
		Render() render.Render
	}

	IResponseStatusRuleFun func(c *gin.Context, response IResponse) int

	JsonResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    any    `json:"data"`
		Error   error  `json:"-"`
	}
)

func ContextIResponse(c *gin.Context, response IResponse, statusRuleFuns ...IResponseStatusRuleFun) {
	if len(statusRuleFuns) > 0 {
		response.WithStatusCode(statusRuleFuns[0](c, response))
	}
	c.Render(response.StatusCode(), response.Render())
	if response.Abort() {
		c.Abort()
	}
}
func JsonErrorString(ctx *gin.Context, err string, codes ...int) {
	JsonError(ctx, fmt.Errorf(err), codes...)
}

func JsonError(ctx *gin.Context, err error, codes ...int) {
	code := http.StatusInternalServerError
	if len(codes) > 0 {
		code = codes[0]
	}
	response := &JsonResponse{Code: code}
	response.SetErr(err)
	ContextIResponse(ctx, response)
}

func JsonSuccess(ctx *gin.Context, data any, messages ...string) {
	statusCode := http.StatusOK
	message := http.StatusText(statusCode)
	if len(messages) > 0 {
		message = messages[0]
	}
	ContextIResponse(ctx, &JsonResponse{Message: message, Code: statusCode, Data: data})
}

// StatusCode 响应状态码
func (j *JsonResponse) StatusCode() int {
	return j.Code
}

// Render 渲染数据
func (j *JsonResponse) Render() render.Render {
	return render.JSON{Data: j}
}

// Abort 是否抛出异常
func (j *JsonResponse) Abort() bool {
	return j.StatusCode() != http.StatusOK
}

// WithStatusCode 携带状态码
func (j *JsonResponse) WithStatusCode(code int) {
	j.Code = code
}

// WithMessage 携带错误消息信息
func (j *JsonResponse) WithMessage(message string) {
	j.Message = message
	if j.Error == nil {
		j.Error = fmt.Errorf(message)
	}
}
func (j *JsonResponse) SetErr(err error) {
	//自定义HTTPError错误
	if hr, ok := j.Error.(*xerror.HTTPError); ok {
		j.WithStatusCode(hr.Code)
	}
	//bind参数校验失败
	if UtTranslator != nil {
		j.WithStatusCode(http.StatusBadRequest)
		err = Translator.ValidateError(err)
	}
	j.Error = err
	j.WithMessage(j.Error.Error())
}
