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
	"os"
	"path/filepath"

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

func getFiles(dir string, messages Messages) (files []*fileinfo, er error) {
	retVal := make([]*fileinfo, 0, 65536)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			messages.AddError("Error getting file info: %s", path)
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

func process(src, dest string, files []*fileinfo, messages Messages) {
	fileCount := len(files)

	for i, file := range files {
		Print("\rProcessed %d out of %d files. %d errors, %d warnings so far.", i, fileCount, messages.GetErrorCount(), messages.GetWarningCount())

		is, err := os.Open(file.path)
		if err != nil {
			messages.AddError("Error opening file: %s: %s", file.path, err.Error())
			continue
		}
		exinfo, err := exif.Decode(is)
		if err != nil {
			messages.AddError("Error reading meta data: %s: %s", file.path, err.Error())
			continue
		}

		dt, err := exinfo.DateTime()
		if err != nil {
			messages.AddWarning("No meta data in the file. Using file modification time: %s: %s", file.path, err.Error())
		} else {
			dt = file.info.ModTime()
		}

		_, filename := filepath.Split(file.path)
		newDir := dt.Format(DestinationDirectoryFormat)
		newPath := filepath.Join(dest, newDir, filename)

		file.newPath = newPath
	}

	Print("\rProcessed %d out of %d files. %d errors, %d warnings so far.\n", fileCount, fileCount, messages.GetErrorCount(), messages.GetWarningCount())
}

func organize(src, dest string) {
	messages := NewMessages()
	files, err := getFiles(src, messages)
	if err != nil {
		Print("Failed to get file list: %s\n", err)
		return
	}
	process(src, dest, files, messages)

	for _, msg := range messages.GetWarnings() {
		Info("%s\n", msg)
	}
	for _, msg := range messages.GetErrors() {
		Print("%s\n", msg)
	}
}
