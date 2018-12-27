# ormgo


### 功能

* 增删改查
* 分页查询
* 查询过滤字段
* 统计记录数
* 软删除

### 简单用法

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

**更详细的用法请看example目录下的文件**

## [说明文档](https://godoc.org/gopkg.in/noxue/ormgo.v1)
