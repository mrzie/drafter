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
	Port int
	// TemplateDir string
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
