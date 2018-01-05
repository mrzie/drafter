package app

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type context struct {
	Req           *http.Request
	Res           http.ResponseWriter
	Session       *sessions.Session
	queries       *queries
	ReqBody       []byte
	authenticated bool
}

type queries struct {
	value map[string]string
}

func (q *queries) GetString(key string) string {
	value, _ := q.value[key]
	return value
}

func (q *queries) GetInt(key string) int {
	str, _ := q.value[key]
	value, _ := strconv.Atoi(str)

	return value
}

func (ctx *context) GetSession() *sessions.Session {
	if ctx.Session == nil {
		// 这里不确定name究竟用在何处，可能是用户端的cookie名称，暂且如此。
		ctx.Session, _ = session_store.Get(ctx.Req, "__u")
	}

	return ctx.Session
}

func (ctx *context) SaveSession() (err error) {
	if ctx.Session == nil {
		// 如果session不存在
		log.Fatal("save sessions which does not exist")
		return nil
	}
	return ctx.Session.Save(ctx.Req, ctx.Res)
}

func (ctx *context) GetQuery() *queries {
	// 不支持带多个值?a=1&a=2
	if ctx.queries == nil {
		value := make(map[string]string)

		qs, _ := url.ParseQuery(ctx.Req.URL.RawQuery)
		for key, values := range qs {
			value[key] = values[0]
		}
		ctx.queries = &queries{value: value}
	}
	return ctx.queries
}

func (ctx *context) GetReqBody() (body []byte, err error) {
	if ctx.ReqBody != nil {
		return ctx.ReqBody, nil
	}
	body, err = ioutil.ReadAll(ctx.Req.Body)
	if err != nil {
		return
	}
	ctx.ReqBody = body
	return
}

func (ctx *context) GetReqStruct(dest interface{}) (err error) {
	body, err := ctx.GetReqBody()
	if err != nil {
		return
	}
	err = json.Unmarshal(body, dest)
	return
}

func (ctx *context) GetVar() map[string]string {
	return mux.Vars(ctx.Req)
}

func (ctx *context) Send(s []byte) {
	ctx.Res.Write(s)
}

func (ctx *context) SendJson(src interface{}) error {
	data, err := json.Marshal(src)
	if err == nil {
		ctx.Res.Write(data)
	}
	return err
}

type simpleMessage struct {
	Code uint16 `json:"code"`
	Msg  string `json:"msg"`
}

func (ctx *context) SendMessage(msg string) error {
	//	0xx - 正确
	return ctx.SendJson(simpleMessage{000, msg})
}

func (ctx *context) Redirect(p string) error {
	url := *ctx.Req.URL
	url.Path = p
	p = url.String()
	ctx.Res.Header().Set("Location", p)
	ctx.Res.WriteHeader(http.StatusMovedPermanently)
	return nil
}