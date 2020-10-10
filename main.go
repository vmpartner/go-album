package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const (
	source = "D:/photo/0 Unsorted"
	target = "D:/photo/1 Autosort"
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

	// Clear code
	re, _ := regexp.Compile(`[^\p{L}\d_]+`)

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
	err = db.Find(&files, "is_copy = false AND stat_size > 250000 AND stat_size < 10000000").Error
	if err != nil {
		panic(err)
	}
	for _, file := range files {

		// Generate target
		fileDate := file.ExifDate
		if fileDate.IsZero() {
			fileDate = file.StatDate
		}

		// Generate dest path
		h := sha1.New()
		h.Write([]byte(file.Path + strconv.Itoa(int(file.StatSize))))
		fileHash := hex.EncodeToString(h.Sum(nil))
		fileName := path.Base(file.Path)
		fileName = strings.ReplaceAll(fileName, path.Ext(file.Path), "")
		fileName = re.ReplaceAllString(fileName, "")
		fileName = strings.ToLower(fmt.Sprintf("%s_%s%s", fileName, fileHash, path.Ext(file.Path)))
		targetFile := fmt.Sprintf("%s/%d/%s/%s", target, fileDate.Year(), months.Replace(fileDate.Month().String()), fileName)
		err := os.MkdirAll(path.Dir(targetFile), 777)
		if err != nil {
			panic(err)
		}

		// Copy file
		_, err = CopyFile(file.Path, targetFile)
		if err != nil {
			panic(err)
		}

		// Update state
		file.IsCopy = true
		err = db.Save(&file).Error
		if err != nil {
			panic(err)
		}
	}
}
