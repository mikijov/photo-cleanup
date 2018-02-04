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
	"io"
	"os"
	"path/filepath"

	"github.com/d4l3k/messagediff"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// Diff compares structs, arrays and strings and provides pretty output.
// Return values are diff which is the difference between values in
// string format meant to be output in the console. equal is boolean flag
// indicating if the values are equal.
func Diff(a, b interface{}) (diff string, equal bool) {
	as, aok := a.(string)
	bs, bok := b.(string)
	if aok && bok {
		// if both a and b are strings, compare them as such
		dmp := diffmatchpatch.New()
		diff := dmp.DiffMain(as, bs, false)
		return dmp.DiffPrettyText(diff), as == bs
	}
	// otherwise compare them as structs
	return messagediff.PrettyDiff(a, b)
}

func CopyDirectory(src, dest string) {
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if path == src {
			return nil
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			panic(err)
		}
		newPath := filepath.Join(dest, rel)

		if info.IsDir() {
			if err := os.Mkdir(newPath, 0777); err != nil {
				panic(err)
			}
		} else {
			in, err := os.Open(path)
			if err != nil {
				panic(err)
			}
			out, err := os.Create(newPath)
			if err != nil {
				panic(err)
			}

			written, err := io.Copy(out, in)
			if err != nil {
				panic(err)
			} else if written != info.Size() {
				panic(fmt.Sprintf("did not write whole file; %d!=%d", written, info.Size()))
			}

			if err := in.Close(); err != nil {
				panic(err)
			}

			if err := out.Chmod(info.Mode() & os.ModePerm); err != nil {
				panic(err)
			}
			if err := out.Close(); err != nil {
				panic(err)
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}
