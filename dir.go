package main

import (
	"github.com/moskvorechie/logs"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"io/ioutil"
	"path"
	"regexp"
	"strings"
	"time"
)

var re2 *regexp.Regexp

func init() {
	re2, _ = regexp.Compile(`\d`)
}

type Dir struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Parent     *Dir     `gorm:"-"`
	DB         *gorm.DB `gorm:"-"`
	Level      int
	Logger     logs.Log `gorm:"-"`
	Path       string   `gorm:"uniqueIndex"`
	ParentPath string
	RootName   string
	LevelName  string
	IsCopy     bool `gorm:"default:false"`
	IsAnalyze  bool `gorm:"default:false"`
}

func (d *Dir) Scan() error {

	parent := &Dir{
		Path: source,
	}
	if d.Parent != nil {
		parent = d.Parent
	}

	// Find dir in DB
	var dir Dir
	err := d.DB.First(&dir, "path = ?", d.Path).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return errors.WithStack(err)
	}

	// Dir
	dir.Path = d.Path
	dir.Level = parent.Level + 1
	dir.ParentPath = parent.Path
	if dir.Level == 2 {
		dir.RootName = strings.TrimSpace(strings.Title(re2.ReplaceAllString(path.Base(dir.Path), "")))
	} else {
		dir.RootName = parent.RootName
	}
	if dir.Level == 3 {
		dir.LevelName = strings.TrimSpace(strings.Title(path.Base(dir.Path)))
	} else {
		dir.LevelName = parent.LevelName
	}

	// Save
	err = d.DB.Save(&dir).Error
	if err != nil {
		return errors.WithStack(err)
	}

	// Scan
	files, err := ioutil.ReadDir(dir.Path)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, f := range files {
		if f.IsDir() {

			d.Logger.InfoF("Read dir %s", dir.Path+"/"+f.Name())

			// Dir
			ds := Dir{
				DB:         d.DB,
				Level:      dir.Level + 1,
				Path:       dir.Path + "/" + f.Name(),
				Parent:     &dir,
				Logger:     d.Logger,
				ParentPath: dir.Path,
			}
			err = ds.Scan()
			if err != nil {
				return errors.WithStack(err)
			}

		} else {

			totalFiles++

			// Save file
			file := File{
				DB:      d.DB,
				Path:    dir.Path + "/" + f.Name(),
				Parent:  &dir,
				DirPath: dir.Path,
				Logger:  d.Logger,
			}
			err = file.Save()
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}
