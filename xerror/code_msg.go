package xerror

import "fmt"

// NewCodeMsg creates a new CodeMsg.
func NewCodeMsg(code int, msg string) *CodeMsg {
	return &CodeMsg{Code: code, Msg: msg}
}

// CodeMsg is a struct that contains a code and a message.
// It implements the error interface.
type CodeMsg struct {
	Code int
	Msg  string
}

func (c *CodeMsg) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", c.Code, c.Msg)
}
