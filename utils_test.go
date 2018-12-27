/**
 * @author 刘荣飞 yes@noxue.com
 * @date 2018/12/27 21:02
 */
package ormgo

import (
	"testing"
)

type MyUser struct {

}

func TestGetCName(t *testing.T) {
	var a interface{}
	a=&[]MyUser{}
	if getCName(a) != "MyUser" {
		t.Error("GetCName测试失败")
	}
}
