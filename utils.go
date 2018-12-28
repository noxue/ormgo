package ormgo

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// 记录哪些结构体模型需要软删除，无需每次都反射结构体
var softDeleteMap map[string]bool

func init() {
	softDeleteMap = make(map[string]bool)
}

// 获取类名字
func getCName(class interface{}) string {
	if class == nil {
		return ""
	}
	name := reflect.TypeOf(class).String()
	arr := strings.Split(name, ".")
	return arr[len(arr)-1]
}

// 检查指定的值是否时nil
// 如果时nil，执行panic
// 调用者需要用recover处理错误
func isNil(doc interface{}) {
	if doc == nil {
		CheckErr(errors.New("doc不能为nil，请先调用SetDoc"))
	}
}

// 判断是否需要软删除
// 记录了调用UseSoftDelete的模型
func needSoftDelete(structName interface{}) (ok bool) {
	name := getCName(structName)
	_, ok = softDeleteMap[name]
	return
}

// 哪些模型需要软删除
func UseSoftDelete(docs ...interface{}) {
	for _, doc := range docs {
		_, ok := softDeleteMap[getCName(doc)]
		if ok {
			CheckErr(errors.New("UseSoftDelete 无需重复调用"))
		}
		softDeleteMap[getCName(doc)] = true
	}
}

var (
	zeroVal  reflect.Value
	zeroArgs []reflect.Value
)

// 调用模型的指定方法
func callToDoc(method string, doc interface{}) error {
	docV := reflect.ValueOf(doc)
	if docV.Kind() != reflect.Ptr {
		e := fmt.Sprintf("ormgo: Passed non-pointer: %v (kind=%v), method:%s", doc, docV.Kind(), method)
		return errors.New(e)
	}
	fn := docV.Elem().Addr().MethodByName(method)
	if fn != zeroVal {
		ret := fn.Call(zeroArgs)
		if len(ret) > 0 && !ret[0].IsNil() {
			return ret[0].Interface().(error)
		}
	}
	return nil
}
