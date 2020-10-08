package main

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"io/ioutil"
	"time"
)

type Dir struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Parent     *Dir     `gorm:"-"`
	DB         *gorm.DB `gorm:"-"`
	Level      int
	Path       string `gorm:"uniqueIndex"`
	ParentPath string
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
	err := d.DB.First(&dir, "path = ?", parent.Path+"/"+d.Path).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return errors.WithStack(err)
	}

	// Dir
	dir.Path = d.Path
	dir.Level = parent.Level + 1
	dir.ParentPath = parent.Path

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
				break
			}

		}

	}

	return nil
}
