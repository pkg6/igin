package xhttp

import (
	"encoding/xml"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg6/igin"
	"github.com/pkg6/igin/xerror"
	"net/http"
)

var (
	xmlVersion  = "1.0"
	xmlEncoding = "UTF-8"
	// BusinessCodeOK represents the business code for success.
	BusinessCodeOK = 0
	// BusinessMsgOk represents the business message for success.
	BusinessMsgOk = "ok"
	// BusinessCodeError represents the business code for error.
	BusinessCodeError = -1
)

type BaseResponse[T any] struct {
	// Code represents the business code, not the http status code.
	Code int `json:"code" xml:"code"`
	// Msg represents the business message, if Code = BusinessCodeOK,
	// and Msg is empty, then the Msg will be set to BusinessMsgOk.
	Msg string `json:"msg" xml:"msg"`
	// Data represents the business data.
	Data T `json:"data,omitempty" xml:"data,omitempty"`
}

type baseXmlResponse[T any] struct {
	XMLName  xml.Name `xml:"xml"`
	Version  string   `xml:"version,attr"`
	Encoding string   `xml:"encoding,attr"`
	BaseResponse[T]
}

// JsonBaseResponse writes v into w with http.StatusOK.
//
func JsonBaseResponse(c *gin.Context, v any) {
	//使用c.ShouldBind() 响应代码才能是200，否则就被gin拦截响应400
	c.JSON(http.StatusOK, wrapBaseResponse(v))
}

// XmlBaseResponse writes v into w with http.StatusOK.
func XmlBaseResponse(c *gin.Context, v any) {
	//使用c.ShouldBind() 响应代码才能是200，否则就被gin拦截响应400
	c.XML(http.StatusOK, wrapXmlBaseResponse(v))
}

func wrapXmlBaseResponse(v any) baseXmlResponse[any] {
	base := wrapBaseResponse(v)
	return baseXmlResponse[any]{
		Version:      xmlVersion,
		Encoding:     xmlEncoding,
		BaseResponse: base,
	}
}
func wrapBaseResponse(v any) BaseResponse[any] {
	var resp BaseResponse[any]
	switch data := v.(type) {
	case *xerror.HTTPError:
		resp.Code = data.Code
		resp.Msg = fmt.Sprintf("%v", data.Message)
	case xerror.HTTPError:
		resp.Code = data.Code
		resp.Msg = fmt.Sprintf("%v", data.Message)
	case xerror.CodeMsg:
		resp.Code = data.Code
		resp.Msg = fmt.Sprintf("%v", data.Msg)
	case *xerror.CodeMsg:
		resp.Code = data.Code
		resp.Msg = fmt.Sprintf("%v", data.Msg)
	case error:
		resp.Code = BusinessCodeError
		if validationErrors, ok := v.(validator.ValidationErrors); ok && igin.UtTranslator != nil {
			var errs igin.ValidateErrors
			for key, value := range validationErrors.Translate(igin.UtTranslator) {
				errs = append(errs, &igin.ValidateError{
					Key:     key,
					Message: value,
				})
			}
			//参数验证失败
			resp.Msg = errs.Error()
		} else {
			resp.Msg = data.Error()
		}
	default:
		resp.Code = BusinessCodeOK
		resp.Msg = BusinessMsgOk
		resp.Data = v
	}
	return resp
}
