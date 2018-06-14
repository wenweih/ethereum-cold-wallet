package main

import (
	"bytes"
	"os"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	log "github.com/sirupsen/logrus"
)

// SubAddress 监听地址
type SubAddress struct {
	gorm.Model
	Address string `gorm:"type:varchar(42);not null;unique_index"`
}

type ormBbAlias struct {
	*gorm.DB
}

func dbConn() *gorm.DB {
	w := bytes.Buffer{}
	w.WriteString(config.DB)
	w.WriteString("?charset=utf8&parseTime=True")
	dbInfo := w.String()
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
	addressPath := strings.Join([]string{HomeDir(), "eth_address.csv"}, "/")
	addressFile, err := os.OpenFile(addressPath, os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer addressFile.Close()

	addresses := []*csvAddress{}
	if err := gocsv.UnmarshalFile(addressFile, &addresses); err != nil {
		log.Fatalln(err.Error())
	}

	for _, address := range addresses {
		subAddress := SubAddress{
			Address: address.Address,
		}
		db.Where(csvAddress{Address: address.Address}).Attrs(csvAddress{Address: address.Address}).FirstOrCreate(&subAddress)
	}
	log.Info("csv2db done")
}
