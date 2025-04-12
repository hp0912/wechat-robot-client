package robot

import "strconv"

type Ret int

const (
	_Ret_name_0 = "ticket error"
	_Ret_name_1 = "logic errorsys error"
	_Ret_name_2 = "param error"
	_Ret_name_3 = "failed login warnfailed login checkcookie invalid"
	_Ret_name_4 = "login environmental abnormality"
	_Ret_name_5 = "operate too often"
)

var (
	_Ret_index_1 = [...]uint8{0, 11, 20}
	_Ret_index_3 = [...]uint8{0, 17, 35, 49}
)

func (i Ret) String() string {
	switch {
	case i == -14:
		return _Ret_name_0
	case -2 <= i && i <= -1:
		i -= -2
		return _Ret_name_1[_Ret_index_1[i]:_Ret_index_1[i+1]]
	case i == 1:
		return _Ret_name_2
	case 1100 <= i && i <= 1102:
		i -= 1100
		return _Ret_name_3[_Ret_index_3[i]:_Ret_index_3[i+1]]
	case i == 1203:
		return _Ret_name_4
	case i == 1205:
		return _Ret_name_5
	default:
		return "Ret(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}

const (
	ticketError         Ret = -14  // ticket error
	logicError          Ret = -2   // logic error
	sysError            Ret = -1   // sys error
	paramError          Ret = 1    // param error
	failedLoginWarn     Ret = 1100 // failed login warn
	failedLoginCheck    Ret = 1101 // failed login check
	cookieInvalid       Ret = 1102 // cookie invalid
	loginEnvAbnormality Ret = 1203 // login environmental abnormality
	optTooOften         Ret = 1205 // operate too often
)

// BaseResponse 大部分返回对象都携带该信息
type BaseResponse struct {
	Ret    Ret `json:"ret"`
	ErrMsg struct {
		String string `json:"string"`
	} `json:"errMsg"`
}

func (b BaseResponse) Ok() bool {
	return b.Ret == 0
}

func (b BaseResponse) Err() error {
	if b.Ok() {
		return nil
	}
	return b.Ret
}
