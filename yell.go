/*	Copyright (c) 2021, Serhat Şevki Dinçer.
	This Source Code Form is subject to the terms of the Mozilla Public
	License, v. 2.0. If a copy of the MPL was not distributed with this
	file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package yell is yet another minimalistic logging library. It provides four severity
// levels, simple API, sync.Locker support, package-specific loggers, customizations.
package yell

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Severity is log severity type
type Severity uint32

// log severity levels
const (
	Sinfo Severity = iota
	Swarn
	Serror
	Sfatal
	Snolog // disables logging
)

// Sname is the list of severity names (in increasing severity) that appear in logs
var Sname = [...]string{"info:", "warn:", "error:", "fatal:"}

// TimeFormat in logs
var TimeFormat = "2006-01-02 15:04:05.000000"

// UTC allows printing coordinated universal time (instead of local time) in logs
var UTC = false

// Logger provides logging service to packages and applications. Designed use case:
//  package mypkg
//
//  import (
//  	"os"
//  	"github.com/jfcg/yell"
//  )
//
//  // log to stdout with warn or higher severity (for example). Update Logger.Slevel to
//  // change log severity level or disable logging. Message list given to Info() etc.
//  // must not be empty or end with a newline.
//  var Logger = yell.Logger{": mypkg:", os.Stdout, yell.Swarn}
//
//  // Info tries to log message list with info severity
//  func Info(msg ...interface{}) error {
//  	return Logger.Log(yell.Sinfo, msg...)
//  }
//
//  // Warn tries to log message list with warn severity
//  func Warn(msg ...interface{}) error {
//  	return Logger.Log(yell.Swarn, msg...)
//  }
//
//  // Error tries to log message list with error severity
//  func Error(msg ...interface{}) (err error) {
//  	err = Logger.Log(yell.Serror, msg...)
//  	// extra stuff for error severity
//  	return
//  }
//
//  // Fatal tries to log message list with fatal severity and panics
//  func Fatal(msg ...interface{}) (err error) {
//  	err = Logger.Log(Sfatal, msg...)
//  	pm := Logger.Package[2:] + yell.Sname[yell.Sfatal]
//  	if err != nil {
//  		pm += err.Error()
//  	}
//  	panic(pm)
//  }
type Logger struct {
	// Package or application name. It must be of the form ": name:"
	Package string

	// Writer is used to log messages, can also be sync.Locker. It must not be nil.
	Writer io.Writer

	// Slevel is minimum severity for logging
	Slevel Severity
}

// for not importing sync
type locker interface {
	Lock()
	Unlock()
}

// Log records message list to Logger if level is severe enough for Logger and the list
// is not empty. Message list must not end with a newline. Log tries to include request
// location (file.go:line) in records, so it must be called as described in Logger doc.
// If Logger.Writer also implements sync.Locker, Lock/Unlock is used to protect logging.
func (lg *Logger) Log(level Severity, msg ...interface{}) (err error) {

	if !(len(msg) > 0 && lg.Slevel <= level && level < Snolog) {
		return // empty msg or ignored level
	}

	// prepare all input to Fprintln before possible locking
	now := time.Now()
	if UTC {
		now = now.UTC()
	}
	prem := now.Format(TimeFormat) + lg.Package + Sname[level]

	// try to locate request location
	_, file, line, ok := runtime.Caller(2)
	if ok {
		file = filepath.Base(file) // full path to file name
		prem += fmt.Sprintf(" %s:%d:", file, line)
	}

	// prepend prem
	msg = append([]interface{}{prem}, msg...)

	// see if Writer is also a sync.Locker
	if lc, ok := lg.Writer.(locker); ok {

		lc.Lock() // lock just before logging
		defer lc.Unlock()
	}

	fmt.Fprintln(lg.Writer, msg...)
	return
}

// Default logger
var Default = Logger{
	// application name
	": " + filepath.Base(os.Args[0]) + ":",

	// log to standard output with warn or higher severity
	os.Stdout, Swarn}

// Info tries to log message list with info severity to Default logger
func Info(msg ...interface{}) error {
	return Default.Log(Sinfo, msg...)
}

// Warn tries to log message list with warn severity to Default logger
func Warn(msg ...interface{}) error {
	return Default.Log(Swarn, msg...)
}

// Error tries to log message list with error severity to Default logger
func Error(msg ...interface{}) error {
	return Default.Log(Serror, msg...)
}

// Fatal tries to log message list with fatal severity to Default logger and panics
func Fatal(msg ...interface{}) (err error) {
	err = Default.Log(Sfatal, msg...)
	pm := Default.Package[2:] + Sname[Sfatal]
	if err != nil {
		pm += err.Error()
	}
	panic(pm)
}
