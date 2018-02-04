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
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xor-gate/goexif2/exif"
)

var DestinationDirectoryFormat string
var MinSize int64
var AllFiles bool
var HiddenFiles bool
var UseExifTime bool
var UseFileTime bool
var UseFilenameEncodedTime bool
var DryRun bool

var FilenameWithTimeRE = regexp.MustCompile("^(?i:IMG|VID)_([[:digit:]]{8}_[[:digit:]]{6}).(?i:jpg|mp4)$")
var TimeLayoutFromFilenameWithDate = TimeFormat("yyyymmdd_HHMMSS")

var acceptedFileTypes = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	// ".mp4":  true,
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
	PreRun: func(cmd *cobra.Command, args []string) {
		DestinationDirectoryFormat = TimeFormat(DestinationDirectoryFormat)
	},
	Run: func(cmd *cobra.Command, args []string) {
		organize(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(organizeCmd)

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// organizeCmd.PersistentFlags().String("foo", "", "A help for foo")

	organizeCmd.Flags().StringVar(&DestinationDirectoryFormat, "dir-fmt", "yyyy/mm", "Directory format")
	organizeCmd.Flags().Int64Var(&MinSize, "min-size", 0, "Minimum file size to consider for processing.")
	organizeCmd.Flags().BoolVar(&AllFiles, "all-files", false, "Process all files. Default is only images (jpg).")
	organizeCmd.Flags().BoolVar(&HiddenFiles, "hidden-files", false, "Process hidden files. Default is only normal files.")
	organizeCmd.Flags().BoolVar(&UseExifTime, "use-exif-time", true, "Use time from exif meta data.")
	organizeCmd.Flags().BoolVar(&UseFileTime, "use-file-time", false, "Use file modification time when no meta data.")
	organizeCmd.Flags().BoolVar(&UseFilenameEncodedTime, "use-filename-encoded-time", true, "Attempt to parse time from filename.")
	organizeCmd.Flags().BoolVarP(&DryRun, "dry-run", "n", false, "Do not make any changes to files, only show what would happen.")
}

type fileinfo struct {
	path    string
	newDir  string
	newPath string
	info    os.FileInfo
	message string
	time    time.Time
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

func evaluate(files []*fileinfo, dest string) {
	fileCount := len(files)

	for i, file := range files {
		Print("\rEvaluated %d out of %d files.", i, fileCount)

		foundTime := false

		if !foundTime && UseExifTime {
			is, err := os.Open(file.path)
			if err != nil {
				Info("%s: error opening file (%s)", file.path, err)
			} else {
				exinfo, err := exif.Decode(is)
				if err != nil {
					Info("%s: error reading meta data (%s)", file.path, err)
				} else {
					time, err := exinfo.DateTime()
					if err == nil {
						foundTime = true
						file.time = time
					}
				}
				is.Close()
			}
		}

		if !foundTime && UseFilenameEncodedTime {
			match := FilenameWithTimeRE.FindStringSubmatch(file.info.Name())
			if match != nil {
				time, err := time.Parse(TimeLayoutFromFilenameWithDate, match[1])
				if err == nil {
					foundTime = true
					file.time = time
				}
			}
		}

		if !foundTime && UseFileTime {
			file.time = file.info.ModTime()
			foundTime = true
		}

		if !foundTime {
			file.message = fmt.Sprintf("%s: could not determine date/time", file.path)
			Print("\r%s\n", file.message)
			continue
		}

		newDir := file.time.Format(DestinationDirectoryFormat)
		file.newDir = filepath.Join(dest, newDir)
		file.newPath = filepath.Join(file.newDir, file.info.Name())
	}

	sort.Slice(files, func(i, j int) bool {
		// prioritize older files, first by path/meta
		l, r := files[i], files[j]
		if l.newPath != r.newPath {
			return l.newPath < r.newPath
		}
		// then by file mod time
		if !l.time.Equal(r.time) {
			return l.time.Before(r.time)
		}
		// then larger files
		return l.info.Size() > r.info.Size()
	})

	// find photos that generate same newPath and mark them as duplicates
	var prevFile *fileinfo
	for _, file := range files {
		if prevFile != nil && file.newPath != "" {
			if file.newPath == prevFile.newPath {
				file.newDir = ""
				file.newPath = ""
				file.message = fmt.Sprintf("%s: duplicate: %s", file.path, prevFile.newPath)
				Print(file.message)
			} else {
				prevFile = file
			}
		} else {
			prevFile = file
		}
	}

	Print("\rEvaluated %d out of %d files.\n", fileCount, fileCount)
}

func execute(files []*fileinfo) {
	fileCount := len(files)

	for i, file := range files {
		Print("\rMoved %d out of %d files.", i, fileCount)

		if file.newPath == "" {
			continue
		}

		// guard against overwriting
		dest, err := os.Lstat(file.newPath)
		if err != nil {
			if os.IsNotExist(err) {
				// all is good, proceed
			} else {
				file.message = fmt.Sprintf("%s: problem checking destination: %s", file.newPath, err)
				Print("%s\n", file.message)
				continue
			}
		} else if os.SameFile(file.info, dest) {
			file.message = fmt.Sprintf("%s: same file", file.newPath)
			Print("%s\n", file.message)
			continue
		} else {
			file.message = fmt.Sprintf("%s: already exists", file.newPath)
			Print("%s\n", file.message)
			continue
		}

		if DryRun {
			file.message = fmt.Sprintf("mv %s %s", file.path, file.newPath)
			Print("%s\n", file.message)
		} else {
			if err := OS.MkdirAll(file.newDir, 0777); err != nil {
				file.message = fmt.Sprintf("%s: failed to create directory: %s", file.newDir, err)
				Print("%s\n", file.message)
			} else if err := OS.Rename(file.path, file.newPath); err != nil {
				file.message = fmt.Sprintf("%s: failed to copy: %s", file.newPath, err)
				Print("%s\n", file.message)
			}
		}
	}

	Print("\rMoved %d out of %d files.\n", fileCount, fileCount)
}

func organize(src, dest string) {
	files, err := getFiles(src)
	if err != nil {
		Print("Failed to get file list: %s\n", err)
		return
	}
	evaluate(files, dest)
	execute(files)
}
