package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"path"
	"strings"
)

const (
	source = "D:/photo/"
	target = "D:/sorted/"
)

var months *strings.Replacer

func init() {
	months = strings.NewReplacer(
		"January", "1_Январь",
		"February", "2_Февраль",
		"March", "3_Март",
		"April", "4_Апрель",
		"May", "5_Май",
		"June", "6_Июнь",
		"July", "7_Июль",
		"August", "8_Август",
		"September", "9_Сентябрь",
		"October", "10_Октябрь",
		"November", "11_Ноябрь",
		"December", "12_Декабрь")
}

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

	// Copy files
	var files []File
	err = db.Find(&files, `stat_size > 250000 AND mime_type NOT IN (
       'font/ttf',
       'application/pdf',
       'application/zip',
       'application/vnd.ms-excel',
       'application/octet-stream',
       'application/x-rar-compressed',
       'application/rss+xml',
       'application/msword',
       'text/plain; charset=utf-8',
       'text/html; charset=utf-8'
	)`).Error
	if err != nil {
		panic(err)
	}
	for _, file := range files {

		// Create path
		err := os.MkdirAll(path.Dir(file.DestPath), 777)
		if err != nil {
			panic(err)
		}

		// Copy file
		_, err = CopyFile(file.Path, file.DestPath)
		if err != nil {
			panic(err)
		}

		// Check file and size
		destFile, err := os.Stat(file.DestPath)
		if err != nil {
			panic(err)
		}
		if destFile.Size() != file.StatSize {
			panic("size not same")
		}

		// Update state
		file.IsCopy = true
		err = db.Save(&file).Error
		if err != nil {
			panic(err)
		}
	}
}
