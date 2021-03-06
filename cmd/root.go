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
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string
var dryRun bool
var ignorePermissionDenied bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "photo-cleanup",
	Short: "Photo organizer.",
	Long: `photo-cleanup is a photo organizer.

See help for individual commands for more help.`,
	// RETURN CODES:
	//   0 - indicates success without any errors.
	//   1 - means photo-cleanup encountered errors while processing, but at least some work was performed.
	//   2 - indicates complete failure, meaning photo-cleanup could not do any part of the requested work.`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		Print("%s\n", err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "display more information while processing")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "display no information while processing")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "Do not make any changes to files, only show what would happen.")
	rootCmd.PersistentFlags().BoolVarP(&ignorePermissionDenied, "ignore-permission-denied", "", false, "Do not abort when encountering permission denied folders or files.")
	// rootCmd.PersistentFlags().BoolVarP(&WarningsAsErrors, "warnings-as-errors", "w", false, "treat all warnings as errors")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
