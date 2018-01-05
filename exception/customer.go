package exception

// type ErrMessage struct {
// 	Code int    `json:"-"`
// 	Msg  string `json:"error"`
// }

// func (e Exception) ParseError() ErrMessage {
// 	var err ErrMessage

// 	switch {
// 	case e.Code == 0:
// 		err = ErrMessage{500, "未知错误"}
// 	case e.Code == 900:
// 		err = ErrMessage{500, "未知错误"}
// 	case e.Code >= 100 && e.Code < 200:
// 		// 	请求参数错误
// 		err = ErrMessage{400, "请求参数错误"}
// 	case e.Code >= 200 && e.Code < 300:
// 		// 数据库错误
// 		err = ErrMessage{500, "服务器错误"}
// 	case e.Code >= 300 && e.Code < 400:
// 		// 代码错误
// 		err = ErrMessage{500, "服务器错误"}
// 	case e.Code >= 400 && e.Code < 500:
// 		err = ErrMessage{401, "授权异常"}
// 	case e.Code >= 500 && e.Code < 600:
// 		err = ErrMessage{404, "未找到"}
// 	}

// 	return err
// }
