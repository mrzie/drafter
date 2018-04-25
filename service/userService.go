package service

import (
	"drafter/db"
	e "drafter/exception"
	. "drafter/setting"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Id          bson.ObjectId `json:"id" bson:"_id"`
	Sinaid      string        `json:"sinaid"` // sina uid
	Avatar      string        `json:"avatar"`
	Name        string        `json:"name"`
	Since       int64         `json:"since"`
	SinaProfile string        `json:"sina_profile"`
	Level       UserLevel     `json:"level"` // 0 - new user  1 - trust 2 - doubt 3 - block
}

type UserLevel uint8

const (
	USER_UNKNOWN UserLevel = 0
	USER_TRUSTED UserLevel = 1
	USER_DOUBTED UserLevel = 2
	USER_BLOCKED UserLevel = 3
)

type UserInfo struct {
	Id      bson.ObjectId `json:"id" bson:"_id"`
	Name    string        `json:"name"`
	Avatar  string        `json:"avatar"`
	Profile string        `json:"profile", bson:"sina_profile"`
}

type userService struct {
}

var UserService = new(userService)

// 获取用户授权之后新增用户，或用户登录后
func (this *userService) UpsertUser(sinaid string, avatar string, name string, profile string, exclusive bool) (since int64, id bson.ObjectId, err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	// todo
	// 1. 用户第一次授权会走这个逻辑
	// 2. 用户我们需要获得用户的id来处理一下session

	var user *User
	modify := m{
		"avatar":       avatar,
		"name":         name,
		"sina_profile": profile,
	}
	if exclusive {
		modify["since"] = time.Now().Unix()
	}
	_, err = ac.C("users").Find(m{"sinaid": sinaid}).Apply(mgo.Change{
		Upsert:    true,
		ReturnNew: true,
		Update: m{
			"$set": modify,
		},
	}, user)
	if err != nil {
		return
	}

	return user.Since, user.Id, nil
}

func (this *userService) GetUsers(ids []bson.ObjectId) (users []User, err error) {
	users = []User{}
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	err = ac.C("users").Find(m{"_id": m{"$in": ids}}).All(&users)
	return
}

func (this *userService) GetUserInfos(ids []bson.ObjectId) (userInfos []UserInfo, err error) {
	// var userInfos []UserInfo
	userInfos = []UserInfo{}
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	err = ac.C("users").Find(m{"_id": m{"$in": ids}}).All(&userInfos)
	return
}

type SinaAccessToken struct {
	Token string `json:"access_token"`
}

func (this *userService) getSinaToken(code string, uri string) (token string, err error) {
	var (
		t      SinaAccessToken
		values = url.Values{
			"client_id":     []string{Settings.SinaToken.ClientId},
			"client_secret": []string{Settings.SinaToken.SecretId},
			"grant_type":    []string{"authorization_code"},
			"redirect_uri":  []string{uri},
			"code":          []string{code},
		}
		_url = "https://api.weibo.com/oauth2/access_token?" + values.Encode()
	)
	err = getFromApi(_url, &t)
	if err == nil {
		token = t.Token
	}
	return
}

// 根据code请求access_token接口，后获取相应的用户信息
func (this *userService) UserAuthorize(code string, uri string, exclusive bool) (uinfo UserInfo, since int64, err error) {
	sinaAccessToken, err := this.getSinaToken(code, uri)
	if err != nil {
		return
	}
	userInfo, err := this.fetchUserInfo(sinaAccessToken)
	if err != nil {
		return
	}

	since, oid, err := this.UpsertUser(userInfo.Id, userInfo.Avatar, userInfo.Name, userInfo.ProfileUrl, exclusive)
	if err != nil {
		return
	}
	return UserInfo{
		Id:     oid,
		Name:   userInfo.Name,
		Avatar: userInfo.Avatar,
		// Profile: userInfo.ProfileUrl,
	}, since, nil
}

type SinaUserInfo struct {
	Id         string `json:"idstr"`
	Name       string `json:"screen_name"`
	Avatar     string `json:"avatar_large"`
	ProfileUrl string `json:"profile_url"`
}

func (this *userService) fetchUserInfo(token string) (userInfo SinaUserInfo, err error) {
	// var userInfo SinaUserInfo
	err = getFromApi("https://api.weibo.com/2/users/show.json?"+url.Values{"access_token": []string{token}}.Encode(), &userInfo)
	return
}

func getFromApi(url string, result interface{}) (err error) {
	res, err := http.Get(url)
	if err != nil || res.StatusCode >= 300 || res.StatusCode < 200 {
		return e.HttpFail()
	}
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return e.DecodeResponseFail()
	}

	err = json.Unmarshal(resBody, result)
	if err != nil {
		return e.DecodeResponseFail()
	}
	return nil
}

func (this *userService) Verify(id bson.ObjectId, since int64) (user User, err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	// var u User
	err = ac.C("users").FindId(id).One(&user)
	if err != nil {
		return
	}
	if user.Since != since {
		err = e.TokenExpired()
	}
	return
}

func (this *userService) setUserLevel(id bson.ObjectId, level UserLevel) (err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	err = ac.C("users").Update(m{"_id": id}, m{"$set": m{"level": level}})
	return
}

// 审核中
func (this *userService) doubtUser(id bson.ObjectId) (err error) {
	return this.setUserLevel(id, USER_DOUBTED)
}

// 屏蔽
func (this *userService) blockUser(id bson.ObjectId) (err error) {
	return this.setUserLevel(id, USER_BLOCKED)
}

// 未置信状态
func (this *userService) uncertainForUser(id bson.ObjectId) (err error) {
	return this.setUserLevel(id, USER_UNKNOWN)
}

// 置信
func (this *userService) trustUser(id bson.ObjectId) (err error) {
	return this.setUserLevel(id, USER_TRUSTED)
}

// 少有的直接针对用户调用的服务
// 对于一个未置信用户，开始敏感词检查。若检查无误则置信
func (this *userService) TrustUser(id bson.ObjectId) (sensitiveOne *Comment, err error) {
	users, err := this.GetUsers([]bson.ObjectId{id})
	if err != nil {
		return
	}
	if len(users) == 0 {
		err = e.NotFound()
		return
	}

	// 判断用户身份为未置信用户
	if user := users[0]; user.Level != USER_UNKNOWN {
		err = e.UserStateError()
		return
	}

	return CommentService.CheckUserComments(id, 0)
}
