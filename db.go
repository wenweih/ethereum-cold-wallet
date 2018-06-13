package main

import (
	"bytes"
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// SubAddress 监听地址
type SubAddress struct {
	gorm.Model
	Address string `gorm:"type:varchar(100);not null"`
}

type ormBbAlias struct {
	*gorm.DB
}

func dbConn() *gorm.DB {
	w := bytes.Buffer{}
	w.WriteString(config.DB)
	w.WriteString("?charset=utf8&parseTime=True")
	dbInfo := w.String()
	fmt.Println(dbInfo)
	db, err := gorm.Open("mysql", dbInfo)
	if err != nil {
		panic(err)
	}
	return db
}

// DBMigrate 数据库表迁移
func (db ormBbAlias) DBMigrate() {
	db.AutoMigrate(&SubAddress{})
}

func (db ormBbAlias) csv2db() {

}
