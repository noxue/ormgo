/**
 * @author 刘荣飞 yes@noxue.com
 * @date 2018/12/27 21:28
 */
package example

import (
	"encoding/json"
	"fmt"
	"ormgo"
	"testing"
	"time"
)

type User struct {
	ormgo.Model `bson:",inline"`
	Name        string
	Password    string
	DeletedAt   time.Time
}

func (this *User) BeforeSave() (err error) {
	fmt.Println(this)
	this.Name = fmt.Sprint("aaa", time.Now().Unix())
	return
}

func (this *User) AfterSave() (err error) {
	fmt.Println(this)
	return
}

func init() {
	ormgo.Init("127.0.0.1:27017", "test", false, time.Second*30)
	ormgo.UseSoftDelete(User{})
}

func TestSave(t *testing.T) {
	user := &User{
		Name:     fmt.Sprint("你好", time.Now().Unix()),
		Password: "密码",
		//DeletedAt: time.Now().UTC(),
	}
	//user.SetDoc(user)
	//err := user.Save()
	err := ormgo.Save(user)
	if err != nil {
		t.Error(err)
	}
}

func TestSelect(t *testing.T) {
	var users []User
	err := ormgo.FindAll(ormgo.Query{
		SortFields: []string{"-name"},
		Limit:      3,
		Skip:       1,
	}, &users)
	if err != nil {
		t.Error(err)
	}
	bs, _ := json.Marshal(users)
	fmt.Println(string(bs))
}

func TestCount(t *testing.T) {
	var user User
	user.SetDoc(user)

	query := ormgo.Query{}

	// 统计所有
	query.Contain = ormgo.All
	n, err := user.Count(query)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("全部:", n)

	// 默认查询不包含被软删除的
	n, err = user.Count(query)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("默认查询:", n)

	// 统计被软删除的
	query.Contain = ormgo.DeletedOnly
	n, err = user.Count(query)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("已被删除:", n)
}

func TestSoftDelete(t *testing.T) {
	var users []User
	err := ormgo.FindAll(ormgo.Query{
		Contain: ormgo.DeletedOnly,
	}, &users)
	if err != nil {
		t.Error(err)
	}
	for _, v := range users {
		v.SetDoc(v)
		err := v.Remove(ormgo.M{"name": v.Name})
		if err != nil {
			t.Error(err)
		}
	}

	fmt.Println(users)
	bs, _ := json.Marshal(users)
	fmt.Println(string(bs))
}

func TestTrueDelete(t *testing.T) {
	var users []User
	err := ormgo.FindAll(ormgo.Query{
		Contain: ormgo.DeletedOnly,
	}, &users)
	if err != nil {
		t.Error(err)
	}
	for _, v := range users {
		v.SetDoc(v)
		err := v.RemoveTrue(ormgo.M{"name": v.Name})
		if err != nil {
			t.Error(err)
		}
	}
	fmt.Println(users)
	bs, _ := json.Marshal(users)
	fmt.Println(string(bs))
}

func TestUpdate(t *testing.T) {
	var users []User
	err := ormgo.FindAll(ormgo.Query{
		SortFields: []string{"-name"},
		Limit:      3,
		Skip:       1,
	}, &users)
	if err != nil {
		t.Error(err)
	}
	for _, v := range users {
		v.SetDoc(v)
		err = v.Update(ormgo.M{"name": v.Name}, ormgo.M{"name": v.Name + "	哈哈哈"})
		if err == nil {
			continue
		}
		t.Error(err)
	}
}

func TestUpdateId(t *testing.T) {
	u := User{}
	u.SetDoc(u)
	err := u.UpdateId("5c24f058d66e067ba2e80f0e", ormgo.M{
		"name": "中国我爱你",
	})
	if err != nil {
		t.Error(err)
	}
}
