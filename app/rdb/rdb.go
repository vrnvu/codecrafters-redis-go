package rdb

import (
	"os"
	"path/filepath"
)

type File struct {
	Dir        string
	DBFilename string
	db         *os.File
}

func NewFile(dir, dbFilename string) *File {
	return &File{Dir: dir, DBFilename: dbFilename}
}

func (f *File) Open() error {
	path := filepath.Join(f.Dir, f.DBFilename)
	db, err := os.Open(path)
	if err != nil {
		return err
	}
	f.db = db

	return nil
}

func (f *File) Close() error {
	return f.db.Close()
}
