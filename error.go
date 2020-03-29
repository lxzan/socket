package socket

type Error struct {
	Code int64
	Msg  string
}

func (this *Error) Error() string {
	return this.Msg
}

func (this *Error) Wrap(s string) *Error {
	return &Error{
		Code: this.Code,
		Msg:  this.Msg + ": " + s,
	}
}

func NewError(code int64, msg string) *Error {
	return &Error{
		Code: code,
		Msg:  msg,
	}
}

var (
	ERR_ReadMessage   = NewError(0, "read message error")
	ERR_DecodeMessage = NewError(1, "decode message error")
	ERR_Timeout       = NewError(2, "")
)
