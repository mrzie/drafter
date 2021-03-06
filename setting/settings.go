package setting

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var Settings struct {
	DB struct {
		URL      string
		Database string
	}
	Salt struct {
		SessionSalt  string
		PasswordSalt string
		TokenSalt    string
	}
	Common struct {
		DefaultPageSize int
	}
	Port  int
	Qiniu struct {
		// 图片上传使用七牛的服务
		AK     string
		SK     string
		Domain string
	}
	StaticDir string // 静态文件目录
	ICP       string
	SinaToken struct {
		ClientId string
		SecretId string
	}
}

var filePath = "./settings.json"

func read() (err error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	err = json.Unmarshal(bytes, &Settings)

	if Settings.Common.DefaultPageSize == 0 {
		Settings.Common.DefaultPageSize = 15
	}

	return
}

func init() {
	err := read()
	if err != nil {
		log.Fatal(err)
	}
}
