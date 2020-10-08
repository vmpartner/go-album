package main

import (
	"github.com/gabriel-vasile/mimetype"
	"github.com/pkg/errors"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"gorm.io/gorm"
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
	Path      string `gorm:"uniqueIndex"`
	DirPath   string
	MimeType  string
	ExifModel string
	ExifDate  time.Time
}

func (f *File) Save() error {

	// Find dir in DB
	var file File
	err := f.DB.First(&file, "path = ?", f.Path).Error
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

		f, err := os.Open(file.Path)
		if err != nil {
			return errors.WithStack(err)
		}
		defer f.Close()

		x, _ := exif.Decode(f)
		if x != nil {
			file.ExifModel = ExifGet(x, exif.Model)
			file.ExifDate, _ = x.DateTime()
		}
	}

	// Get meta

	// Save
	err = f.DB.Save(&file).Error
	if err != nil {
		return errors.WithStack(err)
	}

	log.Println(file.Path)
	return nil
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
