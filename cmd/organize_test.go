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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAcceptFile(t *testing.T) {
	tests := []struct {
		path        string
		accepted    bool
		reason      string
		MinSize     int64
		AllFiles    bool
		HiddenFiles bool
	}{
		{"../test/exif-20170202.jpg", true, "", 0, false, false},
		{"../test/exif-20170202.jpg", false, "small file", 100000, false, false},
		{"../test/exif-20180101.jpg", true, "", 0, false, false},
		{"../test/exif-20180201.jpg", true, "", 0, false, false},
		{"../test/.hidden-file.jpg", false, "hidden file", 0, false, false},
		{"../test/.hidden-file.jpg", true, "", 0, false, true},
		{"../test/jpg.wrong-extension", false, "not image file", 0, false, false},
		{"../test/jpg.wrong-extension", true, "", 0, true, false},
		{"../test/no-exif.jpg", true, "", 0, false, false},
		{"../test/not-readable.jpg", false, "not readable file", 0, false, false},
		{"../test/symlink.jpg", false, "not regular file", 0, false, false},
	}

	for i, test := range tests {
		minSize = test.MinSize
		allFiles = test.AllFiles
		hiddenFiles = test.HiddenFiles

		info, err := os.Lstat(test.path)
		if err != nil {
			t.Errorf("%d, file not found: %s", i, err)
			continue
		}
		accepted, reason := acceptExifFile(info)
		if accepted != test.accepted {
			t.Errorf("%d, accepted: expected:%v got:%v", i, test.accepted, accepted)
		}
		if reason != test.reason {
			t.Errorf("%d, reason: expected:'%v' got:'%v'", i, test.reason, reason)
		}
	}
}

func mkInfo(path string) os.FileInfo {
	fi, err := os.Lstat(path)
	if err != nil {
		panic("cannot stat test file " + path)
	}
	return fi
}

func TestGetFiles(t *testing.T) {
	var expected = map[string]struct {
		path    string
		newDir  string
		newPath string
		info    os.FileInfo
		message string
	}{
		"../test/exif-20170202.jpg": {"../test/exif-20170202.jpg", "", "", mkInfo("../test/exif-20170202.jpg"), ""},
		"../test/exif-20180101.jpg": {"../test/exif-20180101.jpg", "", "", mkInfo("../test/exif-20180101.jpg"), ""},
		"../test/exif-20180201.jpg": {"../test/exif-20180201.jpg", "", "", mkInfo("../test/exif-20180201.jpg"), ""},
		// "../test/not-readable.jpg":        // should not appear
		"../test/.hidden-file.jpg": {"../test/.hidden-file.jpg", "", "", mkInfo("../test/.hidden-file.jpg"), ""},
		"../test/no-exif.jpg":      {"../test/no-exif.jpg", "", "", mkInfo("../test/no-exif.jpg"), ""},
		// "../test/symlink.jpg":             // should not appear
		"../test/jpg.wrong-extension":       {"../test/jpg.wrong-extension", "", "", mkInfo("../test/jpg.wrong-extension"), ""},
		"../test/duplicate.jpg":             {"../test/duplicate.jpg", "", "", mkInfo("../test/duplicate.jpg"), ""},
		"../test/empty.jpg":                 {"../test/empty.jpg", "", "", mkInfo("../test/empty.jpg"), ""},
		"../test/duplicate/duplicate.jpg":   {"../test/duplicate/duplicate.jpg", "", "", mkInfo("../test/duplicate/duplicate.jpg"), ""},
		"../test/duplicate/duplicate-1.jpg": {"../test/duplicate/duplicate-1.jpg", "", "", mkInfo("../test/duplicate/duplicate-1.jpg"), ""},
		"../test/IMG_20180304_123456.jpg":   {"../test/IMG_20180304_123456.jpg", "", "", mkInfo("../test/IMG_20180304_123456.jpg"), ""},
		"../test/2018-03-04 12.34.56.mp4":   {"../test/2018-03-04 12.34.56.mp4", "", "", mkInfo("../test/2018-03-04 12.34.56.mp4"), ""},
		"../test/VID_20181231_203040.mp4":   {"../test/VID_20181231_203040.mp4", "", "", mkInfo("../test/VID_20181231_203040.mp4"), ""},
	}

	allFiles = true
	hiddenFiles = true
	minSize = 0
	useFileTime = false

	files, err := getFiles("../test", acceptExifFile)
	if err != nil {
		t.Fatalf("unexpectedly could not get files: %s", err)
	}

	for _, file := range files {
		test, found := expected[file.path]
		if !found {
			t.Errorf("unexpected file: %s", file.path)
		} else {
			delete(expected, file.path)
		}

		if test.path != file.path {
			t.Errorf("path expected:%s got:%s", test.path, file.path)
		}
		if test.newPath != file.newPath {
			t.Errorf("newPath expected:%s got:%s", test.newPath, file.newPath)
		}
		if test.message != file.message {
			t.Errorf("message expected:%s got:%s", test.message, file.message)
		}
		if diff, equal := Diff(test.info, file.info); !equal {
			t.Errorf("%s: FileInfo: %s", test.path, diff)
		}
	}

	for test := range expected {
		t.Errorf("expected but not found: %s", test)
	}
}

func TestGetFiles2(t *testing.T) {
	files, err := getFiles("../does-not-exist", acceptExifFile)
	if files != nil {
		t.Error("expected nil as return")
	}
	if err.Error() != "lstat ../does-not-exist: no such file or directory" {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestEvaluate(t *testing.T) {
	var expected = map[string]struct {
		path    string
		newDir  string
		newPath string
		info    os.FileInfo
		message string
	}{
		"../test/exif-20170202.jpg":       {"../test/exif-20170202.jpg", "dest/2017/02", "dest/2017/02/exif-20170202.jpg", mkInfo("../test/exif-20170202.jpg"), ""},
		"../test/exif-20180101.jpg":       {"../test/exif-20180101.jpg", "dest/2018/01", "dest/2018/01/exif-20180101.jpg", mkInfo("../test/exif-20180101.jpg"), ""},
		"../test/exif-20180201.jpg":       {"../test/exif-20180201.jpg", "dest/2018/02", "dest/2018/02/exif-20180201.jpg", mkInfo("../test/exif-20180201.jpg"), ""},
		"../test/.hidden-file.jpg":        {"../test/.hidden-file.jpg", "dest/2018/02", "dest/2018/02/.hidden-file.jpg", mkInfo("../test/.hidden-file.jpg"), ""},
		"../test/no-exif.jpg":             {"../test/no-exif.jpg", "", "", mkInfo("../test/no-exif.jpg"), "../test/no-exif.jpg: could not determine date/time"},
		"../test/jpg.wrong-extension":     {"../test/jpg.wrong-extension", "dest/2017/02", "dest/2017/02/jpg.wrong-extension", mkInfo("../test/jpg.wrong-extension"), ""},
		"../test/duplicate.jpg":           {"../test/duplicate.jpg", "dest/2017/02", "dest/2017/02/duplicate.jpg", mkInfo("../test/duplicate.jpg"), ""},
		"../test/duplicate/duplicate.jpg": {"../test/duplicate/duplicate.jpg", "dest/2017/02", "dest/2017/02/duplicate.jpg", mkInfo("../test/duplicate/duplicate.jpg"), ""},
		"../test/not-readable.jpg":        {"../test/not-readable.jpg", "", "", mkInfo("../test/not-readable.jpg"), "../test/not-readable.jpg: could not determine date/time"},
		"../test/empty.jpg":               {"../test/empty.jpg", "", "", mkInfo("../test/empty.jpg"), "../test/empty.jpg: could not determine date/time"},
		"../test/IMG_20180304_123456.jpg": {"../test/IMG_20180304_123456.jpg", "dest/2018/03", "dest/2018/03/IMG_20180304_123456.jpg", mkInfo("../test/IMG_20180304_123456.jpg"), ""},
		"../test/2018-03-04 12.34.56.mp4": {"../test/2018-03-04 12.34.56.mp4", "dest/2018/03", "dest/2018/03/2018-03-04 12.34.56.mp4", mkInfo("../test/2018-03-04 12.34.56.mp4"), ""},
		"../test/VID_20181231_203040.mp4": {"../test/VID_20181231_203040.mp4", "dest/2018/12", "dest/2018/12/VID_20181231_203040.mp4", mkInfo("../test/VID_20181231_203040.mp4"), ""},
	}
	files := make([]*fileinfo, 0, len(expected))
	for _, test := range expected {
		files = append(files, &fileinfo{
			path: test.path,
			info: test.info,
		})
	}

	useFileTime = false
	destinationDirectoryFormat = TimeFormat("yyyy/mm")

	evaluate(files, "dest")

	for _, file := range files {
		test, found := expected[file.path]
		if !found {
			t.Errorf("unexpected file: %s", file.path)
		} else {
			delete(expected, file.path)
		}

		if test.path != file.path {
			t.Errorf("path expected:%s got:%s", test.path, file.path)
		}
		if test.newPath != file.newPath {
			t.Errorf("newPath expected:%s got:%s", test.newPath, file.newPath)
		}
		if test.message != file.message {
			t.Errorf("message expected:%s got:%s", test.message, file.message)
		}
		if diff, equal := Diff(test.info, file.info); !equal {
			t.Errorf("%s: FileInfo: %s", test.path, diff)
		}
	}

	for test := range expected {
		t.Errorf("expected but not found: %s", test)
	}
}

func TestProcessDuplicates(t *testing.T) {
	files := make([]*fileinfo, 2)
	files[0] = &fileinfo{
		path:    "../test/duplicate.jpg",
		newDir:  "dest/2017/02",
		newPath: "dest/2017/02/duplicate.jpg",
		info:    mkInfo("../test/duplicate.jpg"),
	}
	files[1] = &fileinfo{
		path:    "../test/duplicate/duplicate.jpg",
		newDir:  "dest/2017/02",
		newPath: "dest/2017/02/duplicate.jpg",
		info:    mkInfo("../test/duplicate/duplicate.jpg"),
	}

	useFileTime = false
	processDuplicates(files)

	if files[0].newPath != "dest/2017/02/duplicate.jpg" {
		t.Errorf("%s: unexpected newPath (%s)", files[0].path, files[0].newPath)
	}
	if files[0].message != "" {
		t.Errorf("%s: unexpected message (%s)", files[0].path, files[0].message)
	}

	if files[1].newPath != "" {
		t.Errorf("%s: unexpected message (%s)", files[1].path, files[1].newPath)
	}
	if files[1].message != "../test/duplicate/duplicate.jpg: duplicate: dest/2017/02/duplicate.jpg" {
		t.Errorf("%s: unexpected message (%s)", files[1].path, files[1].message)
	}
}

func TestEvaluateFallbackToFileTime(t *testing.T) {
	var expected = []struct {
		path    string
		newDir  string
		newPath string
		info    os.FileInfo
		message string
	}{
		{"../test/no-exif.jpg", "dest/2018/01", "dest/2018/01/no-exif.jpg", mkInfo("../test/no-exif.jpg"), ""},
	}
	files := make([]*fileinfo, len(expected))
	for i, test := range expected {
		files[i] = &fileinfo{
			path: test.path,
			info: test.info,
		}
	}

	useFileTime = true
	evaluate(files, "dest")

	for i, file := range files {
		if expected[i].path != file.path {
			t.Errorf("path expected:%s got:%s", expected[i].path, file.path)
		}
		if expected[i].newPath != file.newPath {
			t.Errorf("newPath expected:%s got:%s", expected[i].newPath, file.newPath)
		}
		if expected[i].message != file.message {
			t.Errorf("message expected:%s got:%s", expected[i].message, file.message)
		}
		if diff, equal := Diff(expected[i].info, file.info); !equal {
			t.Errorf("%s: FileInfo: %s", expected[i].path, diff)
		}
	}
}

func TestExecute(t *testing.T) {
	expected := []struct {
		path      string
		newDir    string
		newPath   string
		mkdirall  int
		rename    int
		message   string
		mkdirErr  error
		renameErr error
	}{
		{"../test/exif-20180101.jpg", "", "", 0, 0, "", nil, nil},
		{"../test/exif-20180101.jpg", "dest/2018/01", "dest/2018/01/exif-20180101.jpg", 1, 1, "", nil, nil},
		{"../test/exif-20180101.jpg", "../test", "../test/exif-20180101.jpg", 0, 0, "../test/exif-20180101.jpg: same file", nil, nil},
		{"../test/exif-20180101.jpg", "../test", "../test/exif-20180201.jpg", 0, 0, "../test/exif-20180201.jpg: already exists", nil, nil},
		{"../test/exif-20180101.jpg", "/root", "/root/non-existant.tmp", 0, 0, "/root/non-existant.tmp: problem checking destination: lstat /root/non-existant.tmp: permission denied", nil, nil},
		{"../test/exif-20180101.jpg", "/root", "/root/non-existant.tmp", 0, 0, "/root/non-existant.tmp: problem checking destination: lstat /root/non-existant.tmp: permission denied", nil, nil},
		{"../test/exif-20180101.jpg", "dest/2018/01", "dest/2018/01/exif-20180101.jpg", 1, 0, "dest/2018/01: failed to create directory: test", errors.New("test"), nil},
		{"../test/exif-20180101.jpg", "dest/2018/01", "dest/2018/01/exif-20180101.jpg", 1, 1, "dest/2018/01/exif-20180101.jpg: failed to copy: test", nil, errors.New("test")},
	}

	dryRun = false
	renameDuplicates = false

	files := make([]*fileinfo, 1)

	for _, test := range expected {
		files[0] = &fileinfo{
			path:    test.path,
			newDir:  test.newDir,
			newPath: test.newPath,
			info:    mkInfo(test.path),
		}

		OS := initMockOs()
		OS.mkdirall.retval = test.mkdirErr
		OS.rename.retval = test.renameErr

		execute(files)

		if OS.mkdirall.called != test.mkdirall {
			t.Errorf("%s: mkdirall not called (%d)", test.path, OS.mkdirall.called)
		} else if OS.mkdirall.called > 0 {
			if OS.mkdirall.path != test.newDir {
				t.Errorf("%s: mkdirall wrong parameter(%s)", test.path, OS.mkdirall.path)
			}
			if OS.mkdirall.mode != 0777 {
				t.Errorf("%s: mkdirall wrong parameter(%o)", test.path, OS.mkdirall.mode)
			}
		}
		if OS.rename.called != test.rename {
			t.Errorf("%s: rename not called (%d)", test.path, OS.rename.called)
		} else if OS.rename.called > 0 {
			if OS.rename.oldpath != test.path {
				t.Errorf("%s: rename wrong parameter(%s)", test.path, OS.rename.oldpath)
			}
			if OS.rename.newpath != test.newPath {
				t.Errorf("%s: rename wrong parameter(%s)", test.path, OS.rename.newpath)
			}
		}
		if files[0].message != test.message {
			t.Errorf("%s: invalid message \"%s\"", test.path, files[0].message)
		}
	}
}

func TestExecuteDryRun(t *testing.T) {
	expected := []struct {
		path     string
		newPath  string
		mkdirall int
		rename   int
		message  string
	}{
		{"../test/exif-20180101.jpg", "", 0, 0, ""},
		{"../test/exif-20180101.jpg", "dest/2018/01/exif-20180101.jpg", 0, 0, "mv ../test/exif-20180101.jpg dest/2018/01/exif-20180101.jpg"},
		{"../test/exif-20180101.jpg", "../test/exif-20180101.jpg", 0, 0, "../test/exif-20180101.jpg: same file"},
		{"../test/exif-20180101.jpg", "../test/exif-20180201.jpg", 0, 0, "../test/exif-20180201.jpg: already exists"},
		{"../test/exif-20180101.jpg", "/root/non-existant.tmp", 0, 0, "/root/non-existant.tmp: problem checking destination: lstat /root/non-existant.tmp: permission denied"},
	}

	files := make([]*fileinfo, 1)

	dryRun = true
	renameDuplicates = false

	for _, test := range expected {
		_, dir := filepath.Split(test.newPath)
		info, err := os.Lstat(test.path)
		if err != nil {
			panic(err)
		}
		files[0] = &fileinfo{
			path:    test.path,
			newDir:  dir,
			newPath: test.newPath,
			info:    info,
		}

		OS := initMockOs()

		execute(files)

		if OS.mkdirall.called != test.mkdirall {
			t.Errorf("%s: mkdirall not called (%d)", test.path, OS.mkdirall.called)
		} else if OS.mkdirall.called > 0 {
			if OS.mkdirall.path != dir {
				t.Errorf("%s: mkdirall wrong parameter(%s)", test.path, OS.mkdirall.path)
			}
			if OS.mkdirall.mode != 0777 {
				t.Errorf("%s: mkdirall wrong parameter(%o)", test.path, OS.mkdirall.mode)
			}
		}
		if OS.rename.called != test.rename {
			t.Errorf("%s: rename not called (%d)", test.path, OS.rename.called)
		} else if OS.rename.called > 0 {
			if OS.rename.oldpath != test.path {
				t.Errorf("%s: rename wrong parameter(%s)", test.path, OS.rename.oldpath)
			}
			if OS.rename.newpath != test.newPath {
				t.Errorf("%s: rename wrong parameter(%s)", test.path, OS.rename.newpath)
			}
		}
		if files[0].message != test.message {
			t.Errorf("%s: invalid message \"%s\"", test.path, files[0].message)
		}
	}
}

func TestExecuteSingleDuplicateSkip(t *testing.T) {
	files := []*fileinfo{
		&fileinfo{
			path:    "../test/duplicate.jpg",
			newDir:  "../test/duplicate",
			newPath: "../test/duplicate/duplicate.jpg",
			info:    mkInfo("../test/duplicate.jpg"),
		},
	}

	dryRun = false
	renameDuplicates = false

	OS := initMockOs()

	execute(files)

	if OS.mkdirall.called != 0 {
		t.Errorf("mkdirall called (%d)", OS.mkdirall.called)
	}
	if OS.rename.called != 0 {
		t.Errorf("rename called (%d)", OS.rename.called)
	}

	if files[0].newPath != "../test/duplicate/duplicate.jpg" {
		t.Errorf("0: unexpected newPath (%s)", files[0].newPath)
	}
	if files[0].message != "../test/duplicate/duplicate.jpg: already exists" {
		t.Errorf("0: unexpected message (%s)", files[0].message)
	}
}

func TestExecuteSingleDuplicateRename(t *testing.T) {
	files := []*fileinfo{
		&fileinfo{
			path:    "../test/duplicate.jpg",
			newDir:  "../test/duplicate",
			newPath: "../test/duplicate/duplicate.jpg",
			info:    mkInfo("../test/duplicate.jpg"),
		},
	}

	dryRun = false
	renameDuplicates = true

	OS := initMockOs()

	execute(files)

	if OS.mkdirall.called != 1 {
		t.Errorf("mkdirall not called (%d)", OS.mkdirall.called)
	} else if OS.mkdirall.called > 0 {
		if OS.mkdirall.path != "../test/duplicate" {
			t.Errorf("mkdirall wrong parameter (%s)", OS.mkdirall.path)
		}
		if OS.mkdirall.mode != 0777 {
			t.Errorf("mkdirall wrong parameter (%o)", OS.mkdirall.mode)
		}
	}
	if OS.rename.called != 1 {
		t.Errorf("rename not called (%d)", OS.rename.called)
	} else if OS.rename.called > 0 {
		if OS.rename.oldpath != "../test/duplicate.jpg" {
			t.Errorf("rename wrong parameter (%s)", OS.rename.oldpath)
		}
		if OS.rename.newpath != "../test/duplicate/duplicate-2.jpg" {
			t.Errorf("rename wrong parameter (%s)", OS.rename.newpath)
		}
	}

	if files[0].newPath != "../test/duplicate/duplicate-2.jpg" {
		t.Errorf("0: unexpected newPath (%s)", files[0].newPath)
	}
	if files[0].message != "" {
		t.Errorf("0: unexpected message (%s)", files[0].message)
	}
}

func TestExecuteGroup(t *testing.T) {
	expected := []struct {
		path     string
		newDir   string
		newPath  string
		newPath2 string
		message  string
	}{
		// no destination; skip
		{"../test/exif-20180101.jpg", "", "", "", ""},
		// normal
		{"../test/exif-20180101.jpg", "dest/2018/01", "dest/2018/01/exif-20180101.jpg", "dest/2018/01/exif-20180101.jpg", ""},
		// src==dest; same file skip
		{"../test/exif-20180101.jpg", "../test", "../test/exif-20180101.jpg", "../test/exif-20180101.jpg", "../test/exif-20180101.jpg: same file"},
		// destination exists; skip
		{"../test/duplicate.jpg", "../test/duplicate", "../test/duplicate/duplicate.jpg", "../test/duplicate/duplicate-2.jpg", ""},
		// inaccessible destination
		{"../test/exif-20180101.jpg", "/root", "/root/non-existant.tmp", "/root/non-existant.tmp", "/root/non-existant.tmp: problem checking destination: lstat /root/non-existant.tmp: permission denied"},
	}

	files := make([]*fileinfo, len(expected))
	for i, test := range expected {
		files[i] = &fileinfo{
			path:    test.path,
			newDir:  test.newDir,
			newPath: test.newPath,
			info:    mkInfo(test.path),
		}
	}

	dryRun = false
	renameDuplicates = true

	OS := initMockOs()

	execute(files)

	if OS.mkdirall.called != 2 {
		t.Errorf("mkdirall not called (%d)", OS.mkdirall.called)
	}
	if OS.rename.called != 2 {
		t.Errorf("rename not called (%d)", OS.rename.called)
	}

	for i, test := range expected {
		if files[i].newPath != test.newPath2 {
			t.Errorf("%s: invalid newPath \"%s\"", test.path, files[i].newPath)
		}
		if files[i].message != test.message {
			t.Errorf("%s: invalid message \"%s\"", test.path, files[i].message)
		}
	}
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags

	destinationDirectoryFormat = TimeFormat(destinationDirectoryFormat)
	verbose = false
	quiet = true

	mtime, err := time.Parse(time.RFC3339, "2018-01-01T12:00:00Z")
	if err != nil {
		fmt.Printf("Problem setting up test time. (%s)\n", err)
		os.Exit(1)
	}
	if err := os.Chtimes("../test/no-exif.jpg", mtime, mtime); err != nil {
		fmt.Printf("Unable to change test file modification time. (%s)\n", err)
		os.Exit(1)
	}
	if err := os.Chmod("../test/not-readable.jpg", 0200); err != nil {
		fmt.Printf("Unable to change test file permission. (%s)\n", err)
		os.Exit(1)
	}

	retVal := m.Run()

	if err := os.Chmod("../test/not-readable.jpg", 0600); err != nil {
		fmt.Printf("Unable to revert test file permission. (%s)\n", err)
	}

	os.Exit(retVal)
}
