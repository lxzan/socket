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
	ERR_ReadMessage   = NewError(0, "read message exception")
	ERR_DecodeMessage = NewError(0, "decode message exception")
)
