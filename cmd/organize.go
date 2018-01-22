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

func organize(src, dest string) {
	filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		_, filename := filepath.Split(path)
		newDir := info.ModTime().Format(DestinationDirectoryFormat)

		newPath := filepath.Join(dest, newDir, filename)

		fmt.Printf("%s => %s\n", relPath, newPath)
		return nil
	})
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
