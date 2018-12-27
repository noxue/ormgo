package ormgo

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

var (
	// 时间的0值，默认插入数据库没赋值的时间的值
	zeroTime time.Time
)

func init() {
	zeroTime, _ = time.Parse("2006-01-02 15:04:05", "0001-01-01 00:00:00")
}

// 与 bson.M 一样
type M map[string]interface{}

// 被用户模型继承
type Model struct {
	// 被操作的模型，因为有的操作需要通过模型得到表名，所以保存在这里，通过SetDoc方法来设置
	doc interface{} `bson:"-" json:"-" xml:"-"`
}

// 查询内容包含的类型
type ContainType int

const (
	ContainTypeDefault ContainType = iota // 默认只查询没删除的结果，即 DeletedAt 为0的结果
	DeletedOnly                           // 只查询已经删除的
	All                                   // 查询所有，包含已删除的
)

type Query struct {
	// 查询条件，与mgo条件一样
	Condition M
	// 过滤字段，例子 {"name":true,"password":false} true表示显示，false表示过滤
	Selector map[string]bool
	// 排序规则，例子 {"-name","password"} 表示 name字段倒叙，password字段正序
	SortFields []string
	// 要查询的条数，和Skip实现分页
	Limit int
	// 跳过多少条，类似mysql的offset
	Skip int
	// 在使用软删除的时候，控制查询范围，可选值 默认 ContainTypeDefault，DeletedOnly，All
	Contain ContainType
}

// 所有通过用户自定义模型调用model提供的函数之前，都必须先调用此函数，因为需要通过模型名称反射出表名
func (this *Model) SetDoc(doc interface{}) {
	this.doc = doc
}

// 保存单个对象
func (this *Model) Save() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(OrmError)
		}
	}()
	isNil(this.doc)
	err = Save(this.doc)
	return
}

// 批量插入对象
//
// 根据提供的类型插入到不同的文档中
func Save(doc interface{}, docs ...interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(OrmError)
		}
	}()

	session := db.getSession()
	defer session.Close()
	d := session.DB(db.dbName)

	err = session.DB(db.dbName).C(getCName(doc)).Insert(doc)
	CheckErr(err)

	// 批量插入
	for _, v := range docs {
		err = d.C(getCName(v)).Insert(v)
		CheckErr(err)
	}
	return
}

// 批量查询数据，返回到docs中
//
// docs数一个模型数组，用于接收查询到的数据
func FindAll(query Query, docs interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(OrmError)
		}
	}()
	session := db.getSession()
	defer session.Close()

	// 我们用nil表示不加任何查询条件，Find方法传nil进去会报错，所以设置为空M类型
	if query.Condition == nil {
		query.Condition = M{}
	}
	// 处理软删除
	if needSoftDelete(docs) {
		if query.Contain == ContainTypeDefault {
			query.Condition["deletedat"] = zeroTime
		} else if query.Contain == DeletedOnly {
			query.Condition["deletedat"] = M{"$ne": zeroTime}
		}
	}

	// 为方便查询的时候可以不用调用SetDoc，通过取docs类型名称做表名
	c := getCName(docs)
	mgoQuery := session.DB(db.dbName).C(c).Find(query.Condition)

	if query.Skip > 0 {
		mgoQuery.Skip(query.Skip)
	}

	if query.Limit > 0 {
		mgoQuery.Limit(query.Limit)
	}

	if query.Selector != nil && len(query.Selector) > 0 {
		mgoQuery.Select(query.Selector)
	}

	if query.SortFields != nil && len(query.SortFields) > 0 {
		mgoQuery.Sort(query.SortFields...)
	}

	err = mgoQuery.All(docs)
	return
}

// 更新满足条件的一个文档
func (this *Model) Update(selector M, doc M) (err error) {
	_, err = update(this.doc, selector, doc, false)
	return
}

func (this *Model) UpdateId(id string, doc M) (err error) {
	_, err = update(this.doc, id, doc, false)
	return
}

func (this *Model) UpdateAll(selector M, doc M) (info *mgo.ChangeInfo, err error) {

	info, err = update(this.doc, selector, doc, true)
	return
}

func update(collectionType interface{}, selector interface{}, doc M, isUpdateAll bool) (info *mgo.ChangeInfo, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(OrmError)
		}
	}()

	isNil(collectionType)

	session := db.getSession()
	defer session.Close()

	if _, ok := selector.(string); ok {
		err = session.DB(db.dbName).C(getCName(collectionType)).UpdateId(bson.ObjectIdHex(selector.(string)), M{"$set": doc})
	} else if isUpdateAll {
		info, err = session.DB(db.dbName).C(getCName(collectionType)).UpdateAll(selector, M{"$set": doc})
	} else {
		err = session.DB(db.dbName).C(getCName(collectionType)).Update(selector, M{"$set": doc})
	}

	return
}

func FindOne(condition M, selector map[string]bool, doc interface{}) (err error) {
	err = find(condition, selector, doc)
	return
}

func FindById(id string, selector map[string]bool, doc interface{}) (err error) {
	err = find(id, selector, doc)
	return
}

func find(condition interface{}, selector map[string]bool, doc interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(OrmError)
		}
	}()
	session := db.getSession()
	defer session.Close()
	isNil(doc)
	coll := session.DB(db.dbName).C(getCName(doc))
	var query *mgo.Query
	if _, ok := condition.(string); ok {
		query = coll.FindId(bson.ObjectIdHex(condition.(string)))
	} else {
		query = coll.Find(condition)
	}

	if selector != nil && len(selector) > 0 {
		query.Select(selector)
	}

	err = query.One(doc)
	return
}

// 软删除单个文档
func (this *Model) Remove(selector M) (err error) {
	_, err = remove(this.doc, selector, false, false)
	return
}

// 真正删除
func (this *Model) RemoveTrue(selector M) (err error) {
	_, err = remove(this.doc, selector, false, true)
	return
}

// 根据文档Id软删除
func (this *Model) RemoveById(id string) (err error) {
	_, err = remove(this.doc, id, false, false)
	return
}

// 根据文档Id真正删除
func (this *Model) RemoveTrueById(id string) (err error) {
	_, err = remove(this.doc, id, false, true)
	return
}

// 软删除所有满足条件的文档
func (this *Model) RemoveAll(selector M) (info *mgo.ChangeInfo, err error) {
	info, err = remove(this.doc, selector, true, false)
	return
}

// 真正删除所有满足条件的文档
func (this *Model) RemoveAllTrue(selector M) (info *mgo.ChangeInfo, err error) {
	info, err = remove(this.doc, selector, true, true)
	return
}

func remove(doc interface{}, selector interface{}, isDeleteAll, isRealDelete bool) (info *mgo.ChangeInfo, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(OrmError)
		}
	}()

	isNil(doc)

	session := db.getSession()
	defer session.Close()
	coll := session.DB(db.dbName).C(getCName(doc))

	doc1 := M{"deletedat": time.Now().UTC()}
	if _, ok := selector.(string); ok {
		if isRealDelete {
			err = coll.RemoveId(bson.ObjectIdHex(selector.(string)))
		} else {
			err = coll.Update(bson.ObjectIdHex(selector.(string)), M{"$set": doc1})
		}
	} else if isDeleteAll {
		if isRealDelete {
			info, err = coll.RemoveAll(selector)
		} else {
			info, err = coll.UpdateAll(selector, M{"$set": doc1})
		}
	} else {
		if isRealDelete {
			err = coll.Remove(selector)
		} else {
			err = coll.Update(selector, M{"$set": doc1})
		}
	}

	return
}

// 统计满足条件的记录个数
func (this *Model) Count(query Query) (n int, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(OrmError)
		}
	}()

	session := db.getSession()
	defer session.Close()

	isNil(this.doc)

	if query.Condition == nil {
		query.Condition = M{}
	}

	// 处理软删除
	if needSoftDelete(this.doc) {
		if query.Contain == ContainTypeDefault {
			query.Condition["deletedat"] = zeroTime
		} else if query.Contain == DeletedOnly {
			query.Condition["deletedat"] = M{"$ne": zeroTime}
		}
	}

	n, err = session.DB(db.dbName).C(getCName(this.doc)).Find(query.Condition).Count()
	return
}

// 返回mgo的session，用户自己调用不能满足的其他函数
// 这个session是Copy或Clone返回的，不需要再Copy
// 使用后记得Close
func GetSession() *mgo.Session {

	return db.getSession()
}
