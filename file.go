package main

import (
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/pkg/errors"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"gopkg.in/djherbis/times.v1"
	"gorm.io/gorm"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

func init() {
	exif.RegisterParsers(mknote.All...)
}

type File struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Parent    *Dir     `gorm:"-"`
	DB        *gorm.DB `gorm:"-"`
	Path      string   `gorm:"uniqueIndex"`
	DirPath   string
	MimeType  string
	ExifModel string
	ExifDate  time.Time
	StatSize  int64
	StatDate  time.Time
	IsCopy    bool `gorm:"default:false"`
	IsAnalyze bool `gorm:"default:false"`
}

func (f *File) Save() error {

	// Find file in DB
	var file File
	err := f.DB.First(&file, "path = ?", f.Path).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return errors.WithStack(err)
	}
	if file.IsAnalyze {
		return nil
	}

	// Open file
	fs, err := os.Open(f.Path)
	if err != nil {
		return errors.WithStack(err)
	}
	defer fs.Close()

	// File
	file.Path = f.Path
	file.DirPath = f.Parent.Path

	// Mimetype
	mime, err := mimetype.DetectFile(file.Path)
	if err != nil {
		return errors.WithStack(err)
	}
	file.MimeType = mime.String()

	switch file.MimeType {
	case "image/png", "image/jpeg":
		x := ExifDecode(fs)
		if x != nil {
			file.ExifModel = ExifGet(x, exif.Model)
			file.ExifDate, _ = x.DateTime()
		}
	}

	// Get stat
	stat, _ := fs.Stat()
	file.StatSize = stat.Size()

	// Stat date
	t := times.Get(stat)
	if t.HasBirthTime() {
		file.StatDate = t.BirthTime()
	}

	// Save
	file.IsAnalyze = true
	err = f.DB.Save(&file).Error
	if err != nil {
		return errors.WithStack(err)
	}

	log.Println(file.Path)
	return nil
}

func ExifDecode(fs io.Reader) *exif.Exif {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()
	x, _ := exif.Decode(fs)

	return x
}

func ExifGet(x *exif.Exif, field exif.FieldName) string {
	camModel, err := x.Get(field)
	if err != nil || camModel == nil {
		return ""
	}
	res, _ := camModel.StringVal()
	res = strings.Replace(res, `"`, ``, -1)
	return res
}

func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
