package main

import (
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
}

func (f *File) Save() error {

	fs, err := os.Open(f.Path)
	if err != nil {
		return errors.WithStack(err)
	}
	defer fs.Close()

	// Find dir in DB
	var file File
	err = f.DB.First(&file, "path = ?", f.Path).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return errors.WithStack(err)
	}

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
