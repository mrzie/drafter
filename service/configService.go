package service

import (
	"drafter/db"
	. "drafter/setting"
	"encoding/json"
)

type configService struct{}

var ConfigService = new(configService)

type configEntry struct {
	Value []byte
}

func (this *configService) Get(key string) (result []byte, err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	var entry configEntry
	err = ac.C("config").Find(m{"key": key}).One(&entry)

	result = entry.Value
	return
}

func (this *configService) Set(key string, value []byte) error {
	ac, err := db.Access()
	if err != nil {
		return err
	}
	defer ac.Close()

	_, err = ac.C("config").Upsert(m{"key": key}, m{"$set": m{"value": value}})
	return err
}

type BlogPreferenceValue struct {
	SiteName string `json:"siteName"`
	Domain   string `json:"domain"`
	Author   string `json:"author"`
	Intro    string `json:"intro"`
	PageSize int    `json:"pageSize"`
}

var BlogPreference BlogPreferenceValue

// 以后也许会有authorPreference之类的

func (this *configService) GetBlogPreferences() (p BlogPreferenceValue, err error) {
	raw, err := this.Get("blog_preference")
	if err != nil {
		return
	}

	err = json.Unmarshal(raw, &p)
	if err != nil {
		return
	}
	if p.PageSize == 0 {
		p.PageSize = Settings.Common.DefaultPageSize // 默认pagesize
	}
	BlogPreference = p
	return
}

func (this *configService) SetBlogPreferences(b BlogPreferenceValue) (err error) {
	raw, err := json.Marshal(b)
	if err != nil {
		return
	}

	err = this.Set("blog_preference", raw)
	this.GetBlogPreferences()
	return
}

func (this *configService) initBlogPreference() {
	// 若不存在authenticate，初始化密码为1
	_, err := this.Get("blog_preference")
	if err == nil {
		return
	}
	if err.Error() == "not found" {
		update := BlogPreferenceValue{
			SiteName: "我的日志",
			Intro:    "欢迎来到我的日志",
			PageSize: 12,
		}
		newValue, err := json.Marshal(update)

		if err != nil {
			return
		}

		err = this.Set("blog_preference", newValue)
		if err != nil {
			return
		}

		return
	}
}
