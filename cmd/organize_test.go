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
		accepted, reason := acceptFile(info)
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
		"../test/jpg.wrong-extension":     {"../test/jpg.wrong-extension", "", "", mkInfo("../test/jpg.wrong-extension"), ""},
		"../test/duplicate.jpg":           {"../test/duplicate.jpg", "", "", mkInfo("../test/duplicate.jpg"), ""},
		"../test/empty.jpg":               {"../test/empty.jpg", "", "", mkInfo("../test/empty.jpg"), ""},
		"../test/duplicate/duplicate.jpg": {"../test/duplicate/duplicate.jpg", "", "", mkInfo("../test/duplicate/duplicate.jpg"), ""},
		"../test/IMG_20180304_123456.jpg": {"../test/IMG_20180304_123456.jpg", "", "", mkInfo("../test/IMG_20180304_123456.jpg"), ""},
		"../test/VID_20181231_203040.mp4": {"../test/VID_20181231_203040.mp4", "", "", mkInfo("../test/VID_20181231_203040.mp4"), ""},
	}

	allFiles = true
	hiddenFiles = true
	minSize = 0
	useFileTime = false

	files, err := getFiles("../test")
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
	files, err := getFiles("../does-not-exist")
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
		"../test/duplicate/duplicate.jpg": {"../test/duplicate/duplicate.jpg", "dest/2017/02", "", mkInfo("../test/duplicate/duplicate.jpg"), "../test/duplicate/duplicate.jpg: duplicate: dest/2017/02/duplicate.jpg"},
		"../test/not-readable.jpg":        {"../test/not-readable.jpg", "", "", mkInfo("../test/not-readable.jpg"), "../test/not-readable.jpg: could not determine date/time"},
		"../test/empty.jpg":               {"../test/empty.jpg", "", "", mkInfo("../test/empty.jpg"), "../test/empty.jpg: could not determine date/time"},
		"../test/IMG_20180304_123456.jpg": {"../test/IMG_20180304_123456.jpg", "dest/2018/03", "dest/2018/03/IMG_20180304_123456.jpg", mkInfo("../test/IMG_20180304_123456.jpg"), ""},
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
		path     string
		newPath  string
		mkdirall int
		rename   int
		message  string
	}{
		{"../test/exif-20180101.jpg", "", 0, 0, ""},
		{"../test/exif-20180101.jpg", "dest/2018/01/exif-20180101.jpg", 1, 1, ""},
		{"../test/exif-20180101.jpg", "../test/exif-20180101.jpg", 0, 0, "../test/exif-20180101.jpg: same file"},
		{"../test/exif-20180101.jpg", "../test/exif-20180201.jpg", 0, 0, "../test/exif-20180201.jpg: already exists"},
		{"../test/exif-20180101.jpg", "/root/non-existant.tmp", 0, 0, "/root/non-existant.tmp: problem checking destination: lstat /root/non-existant.tmp: permission denied"},
	}

	files := make([]*fileinfo, 1)

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
