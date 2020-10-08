package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
)

const (
	source = "D:/photo/0 Unsorted"
	target = "D:/photo/1 Autosort"
)

func main() {

	// Connect to DB
	db, err := gorm.Open(sqlite.Open("go-album.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	_ = db.AutoMigrate(&File{})
	_ = db.AutoMigrate(&Dir{})

	// Dir
	dir := Dir{
		Level: 0,
		DB:    db,
		Path:  source,
	}
	err = dir.Scan()
	if err != nil {
		log.Printf("%+v\n", err)
	}
}
