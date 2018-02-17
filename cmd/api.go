// Copyright © 2018 Milutin Jovanović jovanovic.milutin@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"io/ioutil"
	"os"
)

// OsInterface encapsulates all functions from os package which make modifications
// in the file-system.
type OsInterface interface {
	MkdirAll(path string, mode os.FileMode) error
	Rename(oldpath, newpath string) error
	Remove(path string) error
	ReadFile(path string) ([]byte, error)
}

// OS points to implementation of OsInterface and is initialized by a call to
// InitProdOs() or initMockOs()
var OS OsInterface

type prodOs struct{}

func (this *prodOs) MkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}

func (this *prodOs) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (this *prodOs) Remove(path string) error {
	return os.Remove(path)
}

func (this *prodOs) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// InitProdOs initializes OsInterface with production version which calls
// corresponding os package functions.
func InitProdOs() {
	OS = &prodOs{}
}

// MkdirAllParams holds values describing mocked calls to MkdirAll.
type MkdirAllParams struct {
	called int
	path   string
	mode   os.FileMode
	retval error
}

// RenameParams holds values describing mocked calls to Rename.
type RenameParams struct {
	called  int
	oldpath string
	newpath string
	retval  error
}

// RemoveParams holds values describing mocked calls to Remove.
type RemoveParams struct {
	called int
	path   string
	retval error
}

// ReadFileParams holds values describing mocked calls to ReadFile.
type ReadFileParams struct {
	called   int
	path     string
	retBytes []byte
	retError error
}

type mockOs struct {
	mkdirall MkdirAllParams
	rename   RenameParams
	remove   RemoveParams
	readfile ReadFileParams
}

func (this *mockOs) MkdirAll(path string, mode os.FileMode) error {
	this.mkdirall.called++
	this.mkdirall.path = path
	this.mkdirall.mode = mode
	return this.mkdirall.retval
}

func (this *mockOs) Rename(oldpath, newpath string) error {
	this.rename.called++
	this.rename.oldpath = oldpath
	this.rename.newpath = newpath
	return this.rename.retval
}

func (this *mockOs) Remove(path string) error {
	this.remove.called++
	this.remove.path = path
	return this.remove.retval
}

func (this *mockOs) ReadFile(path string) ([]byte, error) {
	this.readfile.called++
	this.readfile.path = path
	return this.readfile.retBytes, this.readfile.retError
}

func initMockOs() *mockOs {
	retVal := &mockOs{}
	OS = retVal
	return retVal
}
