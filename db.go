package ormgo

import (
	"errors"
	"gopkg.in/mgo.v2"
	"time"
)

var db *dbm

type dbm struct {
	session *mgo.Session
	dbName string
	isNew bool
}

// 作用：初始化orm
//
// isNew 表示每次查询是否创建新的socket，即通过此参数控制每次session时copy还是clone
//
// true表示copy，会创建新的socket，开销相对大一点，但是读写数据量大的时候不会阻塞
//
// false表示Clone，会尽量复用主socket，减少创建新连接的开销
func Init(connectUrl string, dbName string, isNew bool, timeout time.Duration) (err error){
	if db == nil {
		var session *mgo.Session
		session, err = mgo.DialWithTimeout(connectUrl, timeout*time.Second)
		if err!=nil{
			return
		}
		db = &dbm{
			session:session,
			dbName:dbName,
			isNew:isNew,
		}
		return
	}
	return
}

func (this *dbm)getSession() *mgo.Session {
	if db==nil {
		CheckErr(errors.New("请先调用orm.Init初始化orm"))
	}

	if this.isNew {
		return this.session.Copy()
	}
	return this.session.Clone()
}
