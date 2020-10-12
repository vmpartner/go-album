package main

import (
	"crypto/sha1"
	"crypto/sha256"
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
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var re *regexp.Regexp

func init() {
	exif.RegisterParsers(mknote.All...)
	re, _ = regexp.Compile(`[^\p{L}\d_]+`)
}

type File struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Parent     *Dir     `gorm:"-"`
	DB         *gorm.DB `gorm:"-"`
	Path       string   `gorm:"uniqueIndex"`
	DirPath    string
	MimeType   string
	Hash       string
	DestPath   string
	ExifModel  string
	ExifDate   time.Time
	StatSize   int64
	MainDate   time.Time
	CreateDate time.Time
	ChangeDate time.Time
	ModDate    time.Time
	IsCopy     bool `gorm:"default:false"`
	IsAnalyze  bool `gorm:"default:false"`
}

func (f *File) GeneratePath() string {
	fileDate := f.ExifDate
	if fileDate.IsZero() {
		fileDate = f.MainDate
	}
	h := sha1.New()
	h.Write([]byte(f.Path + strconv.Itoa(int(f.StatSize))))
	fileName := path.Base(f.Path)
	fileName = strings.ReplaceAll(fileName, path.Ext(f.Path), "")
	fileName = re.ReplaceAllString(fileName, "")
	fileName = strings.ToLower(fmt.Sprintf("%s_%s%s", fileName, f.Hash, path.Ext(f.Path)))

	// /target/2005/1_Январь/Автозвук Екб/file.jpg
	targetFile := fmt.Sprintf("%s/%d/%s/%s/%s", target, fileDate.Year(), months.Replace(fileDate.Month().String()), f.Parent.LevelName, fileName)

	return targetFile
}

func (f *File) Save() error {

	// Open file
	fs, err := os.Open(f.Path)
	if err != nil {
		return errors.WithStack(err)
	}
	defer fs.Close()

	// Get file hash
	h := sha256.New()
	if _, err := io.Copy(h, fs); err != nil {
		return errors.WithStack(err)
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))

	// Find file in DB
	var file File
	err = f.DB.First(&file, "hash = ?", hash).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return errors.WithStack(err)
	}
	if file.IsAnalyze {
		return nil
	}
	file.Parent = f.Parent

	// File
	file.Hash = hash
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
	case "application/vnd.ms-powerpoint":
		return nil
	}

	// Get stat
	stat, _ := fs.Stat()
	file.StatSize = stat.Size()

	// Stat date
	t := times.Get(stat)
	if t.HasBirthTime() {
		file.CreateDate = t.BirthTime()
	}
	if t.HasChangeTime() {
		file.ChangeDate = t.ChangeTime()
	}
	file.ModDate = t.ModTime()
	file.MainDate = file.CreateDate
	if !file.ChangeDate.IsZero() && file.ChangeDate.Before(file.CreateDate) {
		file.MainDate = file.ChangeDate
	}
	if !file.ModDate.IsZero() && file.ModDate.Before(file.MainDate) {
		file.MainDate = file.ModDate
	}

	file.DestPath = file.GeneratePath()

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
