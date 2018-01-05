package service

import (
	"encoding/gob"
	"encoding/json"
	"reflect"
)

type m map[string]interface{}

func init() {
	AuthService.initAuthConfig()
	ConfigService.initBlogPreference()

	ConfigService.GetBlogPreferences()
	TagService.UpdateCounterCache()
	BlogService.ResetCount()

	gob.Register(AuthValue{})

}

func UnmarshalValidMap(raw json.RawMessage, tpl interface{}) (result map[string]interface{}, ok bool) {
	defer recover()
	result = make(map[string]interface{})
	nameCollect := make(map[string]json.RawMessage)
	if err := json.Unmarshal(raw, &tpl); err != nil {
		return
	}
	if err := json.Unmarshal(raw, &nameCollect); err != nil {
		return
	}
	typ := reflect.TypeOf(tpl).Elem()
	value := reflect.ValueOf(tpl).Elem()
	for name, _ := range nameCollect {
		// v := value.FieldByName()
		match := false
		for i, num := 0, typ.NumField(); i < num; i++ {
			if typ.Field(i).Tag.Get("json") == name || typ.Field(i).Name == name {
				match = true
				result[name] = value.FieldByName(typ.Field(i).Name).Interface()
				break
			}
		}
		if !match {
			return
		}
	}
	ok = true
	return

}

func mapIncludesOne(target map[string]interface{}, names []string) bool {
	for _, name := range names {
		if _, ok := target[name]; ok {
			return true
		}
	}
	return false
}
