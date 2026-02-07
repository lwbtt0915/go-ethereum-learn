package db

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Init(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&ERC20Event{}, &ERC721Event{})
	return db
}
