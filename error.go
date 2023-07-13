package igin

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin/xerror"
)

// AddStatusError 向gin.error中追加错误
func AddStatusError(c *gin.Context, err error, codes ...int) {
	if err == nil {
		return
	}
	switch errData := err.(type) {
	case *xerror.CodeMsg:
		codes = []int{errData.Code}
	case *xerror.HTTPError:
		codes = []int{errData.Code}
	}
	if len(codes) > 0 {
		c.Status(codes[0])
	}
	_ = c.Error(err)
}
