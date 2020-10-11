package main

import (
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
	Path       string `gorm:"uniqueIndex"`
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

	dev := 0

	// Scan
	files, err := ioutil.ReadDir(dir.Path)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, f := range files {
		if f.IsDir() {

			// Dir
			ds := Dir{
				DB:         d.DB,
				Level:      dir.Level + 1,
				Path:       dir.Path + "/" + f.Name(),
				Parent:     &dir,
				ParentPath: dir.Path,
			}
			err = ds.Scan()
			if err != nil {
				return errors.WithStack(err)
			}

		} else {

			// Save file
			file := File{
				DB:      d.DB,
				Path:    dir.Path + "/" + f.Name(),
				Parent:  &dir,
				DirPath: dir.Path,
			}
			err = file.Save()
			if err != nil {
				return errors.WithStack(err)
			}

			dev++
			if dev >= 100 {
				//break
			}

		}

	}

	//d.IsSync = true
	//err = d.DB.Save(&d).Error
	//if err != nil {
	//	return errors.WithStack(err)
	//}

	return nil
}
