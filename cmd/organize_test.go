package cmd

import (
	"fmt"
	"os"
	"testing"
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
		MinSize = test.MinSize
		AllFiles = test.AllFiles
		HiddenFiles = test.HiddenFiles

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

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
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
