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

import "fmt"

var quiet = false
var verbose = true

// Info is a wrapper for fmt.Printf but is only printed when verbose && !quiet.
func Info(format string, args ...interface{}) {
	if verbose && !quiet {
		fmt.Printf(format, args...)
	}
}

// Print is a wrapper for fmt.Printf but is only printed !quiet.
func Print(format string, args ...interface{}) {
	if !quiet {
		fmt.Printf(format, args...)
	}
}
