package cmd

import (
	"os"
)

type OsInterface interface {
	MkdirAll(path string, mode os.FileMode) error
	Rename(oldpath, newpath string) error
}

var OS OsInterface

type prodOs struct{}

func (this *prodOs) MkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}

func (this *prodOs) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func initProdOs() {
	OS = &prodOs{}
}

type MkdirAllParams struct {
	called int
	path   string
	mode   os.FileMode
	retval error
}

type RenameParams struct {
	called  int
	oldpath string
	newpath string
	retval  error
}

type mockOs struct {
	mkdirall MkdirAllParams
	rename   RenameParams
}

func (this *mockOs) MkdirAll(path string, mode os.FileMode) error {
	this.mkdirall.called += 1
	this.mkdirall.path = path
	this.mkdirall.mode = mode
	return this.mkdirall.retval
}

func (this *mockOs) Rename(oldpath, newpath string) error {
	this.rename.called += 1
	this.rename.oldpath = oldpath
	this.rename.newpath = newpath
	return this.rename.retval
}

func initMockOs() *mockOs {
	retVal := &mockOs{}
	OS = retVal
	return retVal
}
