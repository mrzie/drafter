package app

// import "reflect"

// func struct2map(obj interface{}) map[string]interface{} {
// 	t := reflect.TypeOf(obj)
// 	v := reflect.ValueOf(obj)

// 	var data = make(map[string]interface{})
// 	for i := 0; i < t.NumField(); i++ {
// 		data[t.Field(i).Tag.Get("json")] = v.Field(i).Interface()
// 	}
// 	return data
// }
