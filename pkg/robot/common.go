package robot

type BuiltinString struct {
	String string `json:"string"`
}

type BuiltinBuffer struct {
	Buffer []byte `json:"buffer,omitempty"`
	ILen   int64  `json:"iLen,omitempty"`
}

type CommonRequest struct {
	Wxid string `json:"wxid"`
}

// BaseResponse 大部分返回对象都携带该信息
type BaseResponse struct {
	Ret    int `json:"ret"`
	ErrMsg struct {
		String string `json:"string"`
	} `json:"errMsg"`
}

func (b BaseResponse) Ok() bool {
	return b.Ret == 0
}
