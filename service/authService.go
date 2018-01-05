package service

import (
	"crypto/md5"
	. "drafter/setting"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	e "drafter/exception"
)

type authService struct {
	authCache *AuthValue
}

var AuthService = new(authService)

type AuthValue struct {
	Token string
	Since int64
}
type authInfo struct {
	Password string // storaged after md5()
	Token    string
	Since    int64 // means time Unix
}

func (this *authService) Login(password string, exclusive bool) (result AuthValue, err error) {
	raw, err := ConfigService.Get("authenticate")
	if err != nil {
		return
	}
	var stored = new(authInfo)
	err = json.Unmarshal(raw, stored)
	if err != nil {
		return
	}

	if this.md5(password) != stored.Password {
		err = e.PasswordIncorrect()
		return
	}

	if this.authCache == nil || exclusive {
		update := this.makeToken()
		stored.Token = update.Token
		stored.Since = update.Since

		var newValue []byte
		newValue, err = json.Marshal(stored)
		if err != nil {
			return
		}
		err = ConfigService.Set("authenticate", newValue)
		if err != nil {
			return
		}

		this.authCache = &update
	}

	result = *this.authCache
	return
}

func (this *authService) md5(src string) string {
	h := md5.New()
	h.Write([]byte(src))

	h.Write([]byte(Settings.Salt.PasswordSalt))
	return hex.EncodeToString(h.Sum(nil))
}

func (this *authService) makeToken() (result AuthValue) {
	result.Since = time.Now().Unix()

	h := md5.New()
	h.Write([]byte(strconv.Itoa(int(result.Since))))

	h.Write([]byte(Settings.Salt.TokenSalt))

	result.Token = hex.EncodeToString(h.Sum(nil))
	return
}

// 修改密码需要提供旧密码
func (this *authService) EditPassword(newPassword string, oldPassword string) (result AuthValue, err error) {
	raw, err := ConfigService.Get("authenticate")
	if err != nil {
		return
	}
	var stored *authInfo
	err = json.Unmarshal(raw, stored)
	if err != nil {
		return
	}

	if this.md5(oldPassword) != stored.Password {
		err = e.PasswordIncorrect()
		return
	}

	// 在修改了新的密码之后，
	result = this.makeToken()
	update := authInfo{
		Password: this.md5(newPassword),
		Token:    result.Token,
		Since:    result.Since,
	}

	newValue, err := json.Marshal(update)

	if err != nil {
		return
	}

	err = ConfigService.Set("authenticate", newValue)
	if err != nil {
		return
	}
	this.authCache = &result
	return
}

func (this *authService) Verify(value AuthValue) (err error) {
	if this.authCache == nil {
		var raw []byte
		raw, err = ConfigService.Get("authenticate")
		if err != nil {
			return
		}
		var stored authInfo
		err = json.Unmarshal(raw, &stored)
		if err != nil {
			return
		}

		this.authCache = &AuthValue{
			Token: stored.Token,
			Since: stored.Since,
		}
	}

	if value.Since != this.authCache.Since {
		err = e.TokenExpired()
		return
	}

	if value.Token != this.authCache.Token {
		err = e.Unauthorized()
		return
	}

	return
}

func (this *authService) initAuthConfig() {
	// 若不存在authenticate，初始化密码为1
	result, err := ConfigService.Get("authenticate")
	if err == nil {
		return
	}
	if err.Error() == "not found" {
		t := this.makeToken()
		update := authInfo{
			Password: this.md5("1"),
			Token:    t.Token,
			Since:    t.Since,
		}
		newValue, err := json.Marshal(update)

		if err != nil {
			return
		}

		err = ConfigService.Set("authenticate", newValue)
		if err != nil {
			return
		}

		this.authCache = &t
		return
	}
}
