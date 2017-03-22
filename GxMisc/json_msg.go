package GxMisc

/**
作者： Kyle Ding
模块：json接口
说明：
创建时间：2015-10-30
**/

import (
	"encoding/json"
	"github.com/bitly/go-simplejson"
	"reflect"
)

// arr, _ := js.Get("test").Get("array").Array()
// i, _ := js.Get("test").Get("int").Int()
// ms := js.Get("test").Get("string").MustString()

//MsgToBuf 结构转化为json字符串
func MsgToBuf(msg interface{}) ([]byte, error) {
	b, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return b, nil
}

//BufToMsg json字符串转化为结构
func BufToMsg(buf []byte) (*simplejson.Json, error) {
	return simplejson.NewJson(buf)
}

//JSONToStruct json字符串转化为结构
func JSONToStruct(js *simplejson.Json, u interface{}) {
	dataStruct := reflect.Indirect(reflect.ValueOf(u))
	dataStructType := dataStruct.Type()
	for i := 0; i < dataStructType.NumField(); i++ {
		field := dataStructType.Field(i)
		fieldv := dataStruct.Field(i)
		if field.Type.Kind() == reflect.Int {
			n, _ := js.Get(field.Name).Int()
			fieldv.SetInt(int64(n))
		} else if field.Type.Kind() == reflect.Int32 {
			n, _ := js.Get(field.Name).Int()
			fieldv.SetInt(int64(n))
		} else if field.Type.Kind() == reflect.Uint32 {
			n, _ := js.Get(field.Name).Int()
			fieldv.SetUint(uint64(n))
		} else if field.Type.Kind() == reflect.Uint64 {
			n, _ := js.Get(field.Name).Int64()
			fieldv.SetUint(uint64(n))
		} else if field.Type.Kind() == reflect.Int64 {
			n, _ := js.Get(field.Name).Int64()
			fieldv.SetInt(n)
		} else if field.Type.Kind() == reflect.String {
			s, _ := js.Get(field.Name).String()
			fieldv.SetString(s)
		} else if field.Type.Kind() == reflect.Bool {
			b, _ := js.Get(field.Name).Bool()
			fieldv.SetBool(b)
		} else if field.Type.Kind() == reflect.Interface {
			JSONToStruct(js.Get(field.Name), fieldv)
		}
	}
}
