# ormgo


### 功能

* 增删改查
* 分页查询
* 查询过滤字段
* 统计记录数
* 软删除
* Hooks

### 简单用法

* 安装

    `go get gopkg.in/noxue/ormgo.v1`

* 例子

```go
package main

import (
	"fmt"
	"gopkg.in/noxue/ormgo.v1"
	"time"
)

type User struct {
	ormgo.Model `bson:",inline"`
	Name        string
	Password    string
	DeletedAt   time.Time
}

func main() {
	ormgo.Init("127.0.0.1:27017", "test", false, time.Second*30)
	user := User{
		Name:     fmt.Sprint("不学网管理员", time.Now().Unix()),
		Password: "noxue.com",
	}

    // 批量添加只需要添加多个参数即可，参数类型可以不同，比如同时添加分类和文章
	err := ormgo.Save(user) 
	if err != nil {
		fmt.Println("添加错误：", err)
		return
	}
	fmt.Println("添加成功")
}
```

* Hook

只对插入操作做了hook，分别是 `BeforeSave` 和 `AfterSave` ,用法如下。

给模型添加方法，函数就会在保存的时候执行，例如：

```go
// 数据入库之前执行
func (this *User) BeforeSave(){
	// 插入之前密码加密
	this.password = md5(this.password)
}

// 插入之后执行
func(User) AfterSave(){
	
}
```

* 添加了 `SessionExec` 函数

方便执行其他非查询类的代码，比如，添加索引，删除表等等

* 给编辑操作添加自定义操作，如 $addToSet。

如果是编辑时使用自定义操作，务必注意传递给ormgo.Update开头的函数的第二个参数除了是具体对象类型之外，必须是ormgo.M类型

原因在于model.go中update函数里面 检测自定义操作的代码如下：

```go
// 检测是否是其他自定义操作，通过判断key值是否是$开头
	isSub := false
	// 这里的M就是ormgo.M 如果你使用了例如bson.M，那就会导致错误
	if _, ok := doc.(M); ok {
		for key, _ := range doc.(M) {
			if key[0] == '$' {
				isSub = true
				break
			}
		}
	}
```

**更详细的用法请看example目录下的文件**

## [说明文档](https://godoc.org/gopkg.in/noxue/ormgo.v1)
