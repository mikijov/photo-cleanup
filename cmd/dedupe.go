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
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/spf13/cobra"
)

var emptyFilesAreIdentical bool
var preferredChunkSize int64

// dedupCmd represents the dedup command
var dedupCmd = &cobra.Command{
	Use:   "dedupe path [path...]",
	Short: "Find and delete duplicate files.",
	// 	Long: `A longer description that spans multiple lines and likely contains examples
	// and usage of using your command. For example:
	//
	// Cobra is a CLI library for Go that empowers applications.
	// This application is a tool to generate the needed files
	// to quickly create a Cobra application.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dedupe(args)
	},
}

func init() {
	rootCmd.AddCommand(dedupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dedupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	dedupCmd.Flags().BoolVar(&emptyFilesAreIdentical, "empty-files-are-identical", false, "treat empty files as identical duplicates")
	dedupCmd.Flags().Int64Var(&preferredChunkSize, "chunk-size", 64*1024, "preferred chunk size when comparing files")
}

type dupeList struct {
	files []*fileinfo
}

func newDupeList(fi *fileinfo) *dupeList {
	return &dupeList{
		files: []*fileinfo{fi},
	}
}

func (this *dupeList) add(fi *fileinfo) {
	if len(this.files) >= cap(this.files) {
		temp := make([]*fileinfo, len(this.files), cap(this.files)*2)
		copy(temp, this.files)
		this.files = temp
	}
	this.files = append(this.files, fi)
}

func deleteFile(path string) error {
	if dryRun {
		Print("rm \"%s\"\n", path)
	} else {
		err := OS.Remove(path)
		if err != nil {
			if ignorePermissionDenied && os.IsPermission(err) {
				Print("%s: %s\n", path, err)
			} else {
				return err
			}
		}
	}
	return nil
}

func dedupe(paths []string) error {
	dupes := make(map[int64]*dupeList)
	dupeCount := 0
	for _, path := range paths {
		files, err := getFiles(path, nil)
		if err != nil {
			return err
		}

		// sort files within same path with intention to prefer files without
		// suffixes like "-1" etc.
		sort.Slice(files, func(i, j int) bool {
			lname := files[i].info.Name()
			lext := filepath.Ext(lname)
			lname = lname[:len(lname)-len(lext)]

			rname := files[j].info.Name()
			rext := filepath.Ext(rname)
			rname = rname[:len(rname)-len(rext)]

			return lname < rname
		})

		dupeCount += len(files)
		for _, info := range files {
			if dup, ok := dupes[info.info.Size()]; ok {
				dup.add(info)
			} else {
				dupes[info.info.Size()] = newDupeList(info)
			}
		}
	}

	// estimate how much memory we can use for in memory buffers
	availableMemory, _ := GetAvailableMemory()
	availableMemory = (availableMemory * 9) / 10

	processed := 0
	for size, dupeList := range dupes {
		if len(dupeList.files) > 1 {
			if size == 0 {
				for i := 1; i < len(dupeList.files); i++ {
					if emptyFilesAreIdentical {
						err := deleteFile(dupeList.files[i].path)
						if err != nil {
							return err
						}
					} else {
						Info("# Group:                         \n")
						Info("## \"%s\"\n", dupeList.files[i].path)
					}
				}
			} else {
				err := dedupeWorker(size, dupeList.files, availableMemory)
				if err != nil {
					return err
				}
			}

			processed += len(dupeList.files)
		} else {
			Info("# Group:                         \n")
			Info("## \"%s\"\n", dupeList.files[0].path)
			processed++
		}

		delete(dupes, size)
		Print("Processed %d of %d files.\r", processed, dupeCount)
	}
	Print("Processed %d of %d files.\n", processed, dupeCount)

	return nil
}

func dedupeWorker(size int64, files []*fileinfo, availableMemory int64) error {
	// calculate how much memory for each file can be loaded at once
	maxChunkSize := availableMemory / int64(len(files))
	maxChunkSize -= maxChunkSize % 4096 // round it to 4K
	if maxChunkSize == 0 || maxChunkSize > preferredChunkSize {
		maxChunkSize = preferredChunkSize
	}

	for _, file := range files {
		var err error
		file.file, err = os.Open(file.path)
		if err != nil {
			return err
		}
		defer file.file.Close()
	}

	processedSize := int64(0)
	for processedSize < size {
		chunkSize := size - processedSize
		if chunkSize > maxChunkSize {
			chunkSize = maxChunkSize
		}

		allDifferent := true
		for i, file := range files {
			// ensure file.contents can accept chunkSize bytes
			if file.contents == nil {
				file.contents = make([]byte, chunkSize)
			} else {
				file.contents = file.contents[:chunkSize]
			}

			read, err := file.file.Read(file.contents)
			if err != nil {
				return err
			} else if int64(read) < chunkSize {
				return errors.New(file.path + ": unexpected end of file")
			}

			for file.matchGroup < i {
				if (files[file.matchGroup].matchGroup == file.matchGroup) && reflect.DeepEqual(file.contents, files[file.matchGroup].contents) {
					allDifferent = false
					break // still identical with my matchGroup
				}
				file.matchGroup++
			}
		}

		processedSize += chunkSize

		if allDifferent {
			break // no need to continue checking, all files are different
		}
	}

	// delete all files that are not beginning of a matchGroup, i.e. they are
	// duplicates of the matchGroup leader
	Info("# Group:                         \n")
	for i, file := range files {
		if file.matchGroup != i {
			deleteFile(file.path)
		} else {
			Info("## \"%s\"\n", file.path)
		}
	}

	return nil
}
