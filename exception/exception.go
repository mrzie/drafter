package exception

import "gopkg.in/mgo.v2/bson"

/*
	用Code来区分状态
	0xx - 正确
	1xx - 用户请求数据有误
	2xx - 数据库错误
	3xx - 代码异常
	4xx - 核验异常
	5xx - 未找到
	6xx - 网络异常
	9xx - 无法识别的错误
*/

type Exception struct {
	// Logical bool `json:"-"`
	Code   uint16 `json:"code"`
	Msg    string `json:"msg"`
	Raw    error  `json:"raw",omitempty`
	Remark string `json:"remark",omitempty`
	// Customer *ErrMessage // 反馈给读者的错误信息
}

func (e Exception) Error() string {
	return e.Msg + e.Remark
}

func Aver(err error) error {
	if err == nil {
		return nil
	}

	excp, ok := err.(Exception)
	if ok {
		return excp
	}
	return Exception{Code: 900, Raw: err}
}

// 1xx
func InvalidId(msg string) Exception {
	return Exception{Code: 101, Msg: "Invalid id", Remark: msg}
}

func IdRequired() Exception {
	// 需要id的场景下缺失id
	return Exception{Code: 102, Msg: "Id required."}
}

func InvalidEdition() Exception {
	return Exception{Code: 103, Msg: "Invalid edition"}
}

func ParamsRequired(name string) Exception {
	return Exception{Code: 104, Msg: "Parameters required: ", Remark: name}
}

func NotebookDoNotExist(id bson.ObjectId) Exception {
	return Exception{Code: 105, Msg: "Notebook do not exist: ", Remark: string(id)}
}

func NotebookAlreadyExist(name string) Exception {
	return Exception{Code: 106, Msg: "Notebook already exist: ", Remark: name}
}

func NoteDoNotExist(id bson.ObjectId) Exception {
	return Exception{Code: 107, Msg: "Note do not exist:", Remark: string(id)}
}

func TagDoNotExist(name string) Exception {
	return Exception{Code: 108, Msg: "Tag do not exist: ", Remark: name}
}

func FrequentRequest() Exception {
	return Exception{Code: 109, Msg: "Frequent request."}
}

func UserBlocked() Exception {
	return Exception{Code: 110, Msg: "User Blocked"}
}

func CommentTooShort() Exception {
	return Exception{Code: 111, Msg: "Comment too short"}
}

//2xx
func DBTimeout() error {
	return Exception{Code: 201, Msg: "DB Timeout"}
}

func PoolInvalid() Exception {
	return Exception{Code: 202, Msg: "Pool Invalid"}
}

func ConfigFormatError(key string) Exception {
	return Exception{Code: 203, Msg: "Config format error", Remark: key}
}

//3xx
func TypeError(msg string) error {
	return Exception{Code: 301, Msg: "Type error", Remark: msg}
}

func SinaIdOccupied(id string) error {
	return Exception{Code: 302, Msg: "Sina id occupied.", Remark: id}
}

// 异常数据
func ValueError() error {
	return Exception{Code: 303, Msg: "Value error."}
}

func UserStateError() error {
	return Exception{Code: 304, Msg: "User state error."}
}

// func SessionNotExist() error {
// 	return Exception{Code: 302, Msg: "Session does not exist"}
// }

//4xx
// func InvalidHashSessionId() error {
// 	// hash id不匹配，用户修改了cookie
// 	return Exception{Code: 401, Msg: "Invalid hash session id"}
// }

func SessionNotFound() error {
	return Exception{Code: 401, Msg: "Session not found"}
}

func Unauthorized() error {
	return Exception{Code: 402, Msg: "Unauthorized."}
}

func PasswordIncorrect() error {
	return Exception{Code: 403, Msg: "Password incorrect."}
}

func TokenExpired() error {
	return Exception{Code: 404, Msg: "Token expired."}
}

func OAuthFail(remark string) error {
	return Exception{Code: 405, Msg: "OAuth fail.", Remark: remark}
}

func NotFound() error {
	return Exception{Code: 501, Msg: "Not Found"}
}

// 6xx
// 网络异常
// 601 上传图片错误
func UploadImageFail() error {
	return Exception{Code: 601, Msg: "Upload image fail."}
}

// 返回状态码不为2xx
func HttpFail() error {
	return Exception{Code: 602, Msg: "Fail to access api."}
}

// 解析response body失败
func DecodeResponseFail() error {
	return Exception{Code: 603, Msg: "Fail to decode api response."}
}
