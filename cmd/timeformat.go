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

import "strings"

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
