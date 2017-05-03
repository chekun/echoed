package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"os"
)

var db *gorm.DB

func init() {
	db = nil
}

func NewDb() *gorm.DB {
	if db != nil {
		return db
	}
	var err error
	db, err = gorm.Open(os.Getenv("DB_DRIVER"), os.Getenv("DB_DSN"))
	if err != nil {
		panic(err)
	}
	return db
}
