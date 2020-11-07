package main

import (
	"github.com/moskvorechie/logs"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"path"
	"runtime/debug"
	"strings"
)

const (
	source = "D:/photo"
	target = "D:/sorted"
)

var months *strings.Replacer
var totalFiles int

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

	// Logs
	lgr, err := logs.New(&logs.Config{
		App:      "go-album",
		FilePath: "logs.txt",
		Clear:    true,
	})
	if err != nil {
		panic("failed to connect database")
	}

	// Exit on error
	defer func() {
		if err := recover(); err != nil {
			lgr.Error("Fatal stack: \n" + string(debug.Stack()))
			lgr.FatalF("Recovered Fatal %v", err)
		}
	}()

	// Dir
	if true {
		dir := Dir{
			Level:  0,
			Logger: lgr,
			DB:     db,
			Path:   source,
		}
		err = dir.Scan()
		if err != nil {
			log.Printf("%+v\n", err)
		}
	}

	lgr.InfoF("Total files touched %d", totalFiles)

	// Copy files
	var files []File
	err = db.Find(&files, `is_copy = false`).Error
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
