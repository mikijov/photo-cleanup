// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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

	organizeCmd.Flags().StringVarP(&DestinationDirectoryFormat, "dir-fmt", "d", TimeFormat("yyyy/mm"), "Directory format")
}

type fileinfo struct {
	path    string
	newPath string
	info    os.FileInfo
}

type organizeError struct {
	message  string
	err      error
	fileinfo fileinfo
}

func (this *organizeError) Error() string {
	return fmt.Sprintf(this.message, this.err, this.fileinfo.path, this.fileinfo.newPath)
}

type Messages interface {
	AddError(msg string)
	AddWarning(msg string)

	GetErrorCount() int
	GetErrors() []string

	GetWarningCount() int
	GetWarnings() []string

	GetMessageCount() int
	GetMessages() []string
}

type messages struct {
	errors   []string
	warnings []string
}

func NewMessages() Messages {
	return &messages{
		errors: make([]string, 0, 1024),
	}
}

func (this *messages) AddWarning(err string) {
	if cap(this.warnings) <= len(this.warnings) {
		newList := make([]string, len(this.warnings), cap(this.warnings)*2)
		copy(newList, this.warnings)
		this.warnings = newList
	}
	this.warnings = append(this.warnings, err)
}

func (this *messages) AddError(err string) {
	if cap(this.errors) <= len(this.errors) {
		newList := make([]string, len(this.errors), cap(this.errors)*2)
		copy(newList, this.errors)
		this.errors = newList
	}
	this.errors = append(this.errors, err)
}

func (this *messages) GetErrorCount() int {
	return len(this.errors)
}

func (this *messages) GetErrors() []string {
	return this.errors
}

func (this *messages) GetWarningCount() int {
	return len(this.warnings)
}

func (this *messages) GetWarnings() []string {
	return this.warnings
}

func (this *messages) GetMessageCount() int {
	return len(this.errors) + len(this.warnings)
}

func (this *messages) GetMessages() []string {
	retVal := make([]string, 0, len(this.warnings)+len(this.errors))
	retVal = append(retVal, this.warnings...)
	retVal = append(retVal, this.errors...)
	return retVal
}

func getFiles(dir string, errors Messages) (files []*fileinfo, er error) {
	retVal := make([]*fileinfo, 0, 65536)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errors.AddError(fmt.Sprintf("Error getting file info: %s", path))
			return err
		}
		if info.IsDir() {
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
			fmt.Printf("\rFound %d files.", len(retVal))
		}
		return nil
	})

	fmt.Printf("\rFound %d files.\n", len(retVal))

	if err != nil {
		return nil, err
	}

	return retVal, nil
}

func process(src, dest string, files []*fileinfo, errors Messages) {
	fileCount := len(files)

	for i, file := range files {
		fmt.Printf("\rProcessed %d out of %d files. %d errors so far.", i, fileCount, errors.GetMessageCount())

		is, err := os.Open(file.path)
		if err != nil {
			errors.AddError(fmt.Sprintf("Error opening file: %s: %s", file.path, err.Error()))
			continue
		}
		exinfo, err := exif.Decode(is)
		if err != nil {
			errors.AddError(fmt.Sprintf("Error reading meta data: %s: %s", file.path, err.Error()))
			continue
		}

		dt, err := exinfo.DateTime()
		if err != nil {
			errors.AddWarning(fmt.Sprintf("No meta data in the file. Using file modification time: %s: %s", file.path, err.Error()))
		} else {
			dt = file.info.ModTime()
		}

		_, filename := filepath.Split(file.path)
		newDir := dt.Format(DestinationDirectoryFormat)
		newPath := filepath.Join(dest, newDir, filename)

		file.newPath = newPath
	}

	fmt.Printf("\rProcessing %d out of %d files. %d errors so far.\n", fileCount, fileCount, errors.GetMessageCount())
}

func organize(src, dest string) {
	errors := NewMessages()
	files, err := getFiles(src, errors)
	if err != nil {
		fmt.Printf("Failed to get file list: %s\n", err)
		return
	}
	process(src, dest, files, errors)

	for _, err := range errors.GetMessages() {
		fmt.Printf("%s\n", err)
	}
}

// TimeFormat creates format string for time.Format() using yyyy etc notation.
//
// I prefer setting the format for time.Format() using familiar yyyy, mmm etc.
// notation rather then example based one in Go. In particular this is useful
// when exposing the format to users, e.g. on the command line.
//
// TODO: Allow escaping of characters.
func TimeFormat(format string) string {
	patterns := []struct {
		from string
		to   string
	}{
		{"yyyy", "2006"},
		{"yy", "06"},
		{"mmmm", "January"},
		{"mmm", "Jan"},
		{"mm", "01"},
		{"dddd", "Monday"},
		{"ddd", "Mon"},
		{"dd", "02"},
		{"HHT", "03"},
		{"HH", "15"},
		{"MM", "04"},
		{"SS", "05"},
		{"ss", "05"},
		{"tt", "PM"},
		{"Z", "MST"},
		{"ZZZ", "MST"},
	}

	retVal := format
	for _, pattern := range patterns {
		retVal = strings.Replace(retVal, pattern.from, pattern.to, -1)
	}
	return retVal
}
