package app

import (
	e "drafter/exception"
	"drafter/service"
	"html/template"
	"net/http"

	"gopkg.in/mgo.v2/bson"
)

type OAuthLoginModel struct {
	Success  bool
	UserInfo service.UserInfo
}

var UserAuthCache = make(map[bson.ObjectId]int64)

func UserVerifyController(ctx *context) (err error) {
	var u *UserAuthInfo
	session := ctx.GetSession()
	v, ok := session.Values["user"]
	if ok {
		u, ok = v.(*UserAuthInfo)
	}
	if !ok {
		return e.SessionNotFound()
	}
	user, err := userS.Verify(u.Id, u.Since)
	if err != nil {
		delete(session.Values, "user")
		ctx.SaveSession()
		return
	} else {
		if ctx.context == nil {
			ctx.context = make(map[string]interface{})
		}
		ctx.context["user"] = &user
	}

	return
}

func UserLogoutController(ctx *context) (err error) {
	delete(ctx.GetSession().Values, "user")
	err = ctx.SaveSession()
	if err != nil {
		return
	}
	return ctx.SendMessage("log out success.")
}

type UserAuthInfo struct {
	Id    bson.ObjectId
	Since int64
}

func OAuthLogin(ctx *context, exclusive bool) {
	var m = OAuthLoginModel{}
	code := ctx.GetQuery().GetString("code")
	uinfo, since, err := userS.UserAuthorize(code, ctx.Req.RequestURI, exclusive)
	if err == nil {
		UserAuthCache[uinfo.Id] = since
		ctx.GetSession().Values["user"] = &UserAuthInfo{Id: uinfo.Id, Since: since}
		err = ctx.SaveSession()
	}
	if err == nil {
		m.Success = true
		m.UserInfo = uinfo
	} else {
		m.Success = false
	}

	t, err := template.New("OAuthTpl").Parse("<script>window.opener.OAuthCallback({{.}})</script>")
	if err == nil {
		err = t.Execute(ctx.Res, m)
	}
	if err != nil {
		ctx.SendMessage("fail")
	}
	return
}

func OAuthLoginController(w http.ResponseWriter, req *http.Request) {
	ctx := GetContext(w, req)
	OAuthLogin(ctx, false)
}

func OAuthExclusiveLoginController(w http.ResponseWriter, req *http.Request) {
	ctx := GetContext(w, req)
	OAuthLogin(ctx, true)
}
