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

// import "fmt"

// var WarningsAsErrors = true
//
// type Messages interface {
// 	AddError(msg string, args ...interface{})
// 	AddWarning(msg string, args ...interface{})
//
// 	GetErrorCount() int
// 	GetErrors() []string
//
// 	GetWarningCount() int
// 	GetWarnings() []string
//
// 	GetMessageCount() int
// 	GetMessages() []string
// }
//
// type messages struct {
// 	errors   []string
// 	warnings []string
// }
//
// func NewMessages() Messages {
// 	return &messages{
// 		errors:   make([]string, 0, 1024),
// 		warnings: make([]string, 0, 1024),
// 	}
// }
//
// func (this *messages) AddWarning(msg string, args ...interface{}) {
// 	if WarningsAsErrors {
// 		this.AddError(msg)
// 	}
//
// 	if cap(this.warnings) <= len(this.warnings) {
// 		newList := make([]string, len(this.warnings), cap(this.warnings)*2)
// 		copy(newList, this.warnings)
// 		this.warnings = newList
// 	}
// 	this.warnings = append(this.warnings, fmt.Sprintf(msg, args...))
// }
//
// func (this *messages) AddError(msg string, args ...interface{}) {
// 	if cap(this.errors) <= len(this.errors) {
// 		newList := make([]string, len(this.errors), cap(this.errors)*2)
// 		copy(newList, this.errors)
// 		this.errors = newList
// 	}
// 	this.errors = append(this.errors, fmt.Sprintf(msg, args...))
// }
//
// func (this *messages) GetErrorCount() int {
// 	return len(this.errors)
// }
//
// func (this *messages) GetErrors() []string {
// 	return this.errors
// }
//
// func (this *messages) GetWarningCount() int {
// 	return len(this.warnings)
// }
//
// func (this *messages) GetWarnings() []string {
// 	return this.warnings
// }
//
// func (this *messages) GetMessageCount() int {
// 	return len(this.errors) + len(this.warnings)
// }
//
// func (this *messages) GetMessages() []string {
// 	retVal := make([]string, 0, len(this.warnings)+len(this.errors))
// 	retVal = append(retVal, this.warnings...)
// 	retVal = append(retVal, this.errors...)
// 	return retVal
// }
