package cmd

import (
	"testing"
)

func TestDedepeWorkerEqual(t *testing.T) {
	var expected = []string{
		"../test/duplicate/duplicate.jpg",
		"../test/duplicate/duplicate-1.jpg",
		"../test/exif-20170202.jpg",
	}
	files := make([]*fileinfo, 0, len(expected))
	for _, test := range expected {
		files = append(files, &fileinfo{
			path: test,
			info: mkInfo(test),
		})
	}

	OS := initMockOs()

	if err := dedupeWorker(55513, files, 4096); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if OS.remove.called != 1 {
		t.Errorf("remove called %d number of times", OS.remove.called)
	}
	if OS.remove.path != "../test/duplicate/duplicate-1.jpg" {
		t.Errorf("remove called for wrong file (%s)", OS.remove.path)
	}
}
