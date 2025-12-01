package gzerror

type BizError struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
}

func (e *BizError) Error() string {
	if e == nil {
		return ""
	}
	return e.Msg
}

func NewBizError(code int64, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}
