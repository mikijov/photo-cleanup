// Copyright © 2018 Milutin Jovanović miki@voreni.com
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
	"strings"

	"github.com/spf13/cobra"
	"github.com/xor-gate/goexif2/exif"
)

var DestinationDirectoryFormat string
var MinSize int64
var AllFiles bool
var HiddenFiles bool
var FallbackToFileTime bool

var acceptedFileTypes = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".mp4":  true,
}

// organizeCmd represents the organize command
var organizeCmd = &cobra.Command{
	Use:   "organize srcdir destdir",
	Short: "Moves photos from source into proper destination subdirectory.",
	// 	Long: `A longer description that spans multiple lines and likely contains examples
	// and usage of using your command. For example:
	//
	// Cobra is a CLI library for Go that empowers applications.
	// This application is a tool to generate the needed files
	// to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		organize(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(organizeCmd)

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// organizeCmd.PersistentFlags().String("foo", "", "A help for foo")

	organizeCmd.Flags().StringVar(&DestinationDirectoryFormat, "dir-fmt", TimeFormat("yyyy/mm"), "Directory format")
	organizeCmd.Flags().Int64Var(&MinSize, "min-size", 0, "Minimum file size to consider for processing.")
	organizeCmd.Flags().BoolVar(&AllFiles, "all-files", false, "Process all files. Default is only images (jpg).")
	organizeCmd.Flags().BoolVar(&HiddenFiles, "hidden-files", false, "Process hidden files. Default is only normal files.")
	organizeCmd.Flags().BoolVar(&FallbackToFileTime, "allow-file-time", false, "Allow file time when no meta data.")
}

type fileinfo struct {
	path    string
	newPath string
	info    os.FileInfo
	message string
}

func acceptFile(info os.FileInfo) (accepted bool, reason string) {
	mode := info.Mode()
	if !mode.IsRegular() {
		return false, "not regular file"
	}
	perm := mode.Perm()
	if perm&0400 != 0400 {
		return false, "not readable file"
	}

	filename := info.Name()
	if !HiddenFiles && filename[0] == '.' {
		return false, "hidden file"
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if !AllFiles && !acceptedFileTypes[ext] {
		return false, "not image file"
	}
	if info.Size() < MinSize {
		return false, "small file"
	}
	return true, ""
}

func getFiles(dir string) (files []*fileinfo, er error) {
	retVal := make([]*fileinfo, 0, 65536)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			Print("\r%s: error getting file info: %s\n", path, err)
			return err
		}
		if info.IsDir() {
			return nil
		}

		if ok, reason := acceptFile(info); !ok {
			Info("\r%s: skipping: %s\n", path, reason)
			return nil
		}

		if cap(retVal) <= len(retVal) {
			newFiles := make([]*fileinfo, cap(retVal)*2)
			copy(newFiles, retVal)
			retVal = newFiles
		}
		retVal = append(retVal, &fileinfo{
			path: path,
			info: info,
		})

		if len(retVal)%1000 == 0 {
			Print("\rFound %d files.", len(retVal))
		}
		return nil
	})

	Print("\rFound %d files.\n", len(retVal))

	if err != nil {
		return nil, err
	}

	return retVal, nil
}

func evaluate(src, dest string, files []*fileinfo) {
	fileCount := len(files)

	for i, file := range files {
		Print("\rEvaluated %d out of %d files.", i, fileCount)

		is, err := os.Open(file.path)
		if err != nil {
			file.message = fmt.Sprintf("%s: error opening file (%s)", file.path, err)
			Print("\r%s\n", file.message)
			is.Close()
			continue
		}
		exinfo, err := exif.Decode(is)
		if err != nil {
			file.message = fmt.Sprintf("%s: error reading meta data (%s)", file.path, err)
			Print("\r%s\n", file.message)
			is.Close()
			continue
		}

		dt, err := exinfo.DateTime()
		if err != nil {
			if FallbackToFileTime {
				file.message = fmt.Sprintf("%s: using file modification time (%s)", file.path, err)
				Info("\r%s\n", file.message)
				dt = file.info.ModTime()
			} else {
				file.message = fmt.Sprintf("%s: no date/time meta data (%s)", file.path, err)
				Print("\r%s\n", file.message)
				is.Close()
				continue
			}
		}

		newDir := dt.Format(DestinationDirectoryFormat)
		file.newPath = filepath.Join(dest, newDir, file.info.Name())

		is.Close()
	}

	Print("\rEvaluated %d out of %d files.\n", fileCount, fileCount)
}

func organize(src, dest string) {
	files, err := getFiles(src)
	if err != nil {
		Print("Failed to get file list: %s\n", err)
		return
	}
	evaluate(src, dest, files)

	for _, file := range files {
		if file.newPath != "" {
			Print("mv %s %s\n", file.path, file.newPath)
		}
	}
}
