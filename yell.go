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
//  // log to stdout with warn or higher severity (for example).
//  var Logger = yell.New(": mypkg:", os.Stdout, yell.Swarn)
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
//  	pm := Logger.Name() + yell.Sname[yell.Sfatal]
//  	if err != nil {
//  		pm += err.Error()
//  	}
//  	panic(pm)
//  }
type Logger struct {
	// name of package or application, must be of the form ": mypkg:"
	name string

	// writer is used to log messages, can also be sync.Locker, must not be nil
	writer io.Writer

	// minLevel is minimum severity for logging
	minLevel Severity
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// New creates a Logger with package/application name (must be of the form ": mypkg:"),
// writer to log (which can also implement sync.Locker to protect logging) and minimum
// severity level to log. Panics if arguments are invalid.
func New(name string, writer io.Writer, minLevel Severity) Logger {
	l := len(name) - 1
	if l < 3 || name[0] != ':' || name[1] != ' ' || isSpace(name[2]) ||
		isSpace(name[l-1]) || name[l] != ':' || writer == nil || minLevel > Snolog {
		panic("Invalid arguments to yell.New()")
	}
	return Logger{name, writer, minLevel}
}

// Name of Logger, skipping ": "
func (lg *Logger) Name() string {
	return lg.name[2:]
}

// for not importing sync
type locker interface {
	Lock()
	Unlock()
}

// UpdateWriter tries to update Logger's writer. If both old & new writers implement
// sync.Locker, they must resolve to the same locker. Otherwise UpdateWriter refuses
// to update because old locker could still be in use in Log() calls while we update.
// Returns true on successful update.
func (lg *Logger) UpdateWriter(writer io.Writer) (success bool) {
	if writer == nil {
		return false
	}

	// see if writers are also sync.Locker
	if lc, ok := lg.writer.(locker); ok {

		if lc2, ok := writer.(locker); ok && lc != lc2 {
			return false // different lockers
		}

		lc.Lock()
		defer lc.Unlock()
	}

	lg.writer = writer
	return true
}

// SetMinLevel sets minimum severity level for logging
func (lg *Logger) SetMinLevel(level Severity) {
	if level > Snolog {
		lg.minLevel = Snolog
	} else {
		lg.minLevel = level
	}
}

// Log records message list to Logger if level is severe enough for Logger and the list
// is not empty. Message list must not end with a newline. Log tries to include request
// location (file.go:line) in records, so it must be called as described in Logger doc.
// If Logger.writer also implements sync.Locker, Lock/Unlock is used to protect logging.
func (lg *Logger) Log(level Severity, msg ...interface{}) (err error) {

	if !(len(msg) > 0 && lg.minLevel <= level && level < Snolog) {
		return // empty msg or ignored level
	}

	// prepare all input to Fprintln before possible locking
	now := time.Now()
	if UTC {
		now = now.UTC()
	}
	prem := now.Format(TimeFormat) + lg.name + Sname[level]

	// try to locate request location
	_, file, line, ok := runtime.Caller(2)
	if ok {
		file = filepath.Base(file) // full path to file name
		prem += fmt.Sprintf(" %s:%d:", file, line)
	}

	// prepend prem
	msg = append([]interface{}{prem}, msg...)

	// see if writer is also a sync.Locker
	if lc, ok := lg.writer.(locker); ok {

		lc.Lock() // lock just before logging
		defer lc.Unlock()
	}

	fmt.Fprintln(lg.writer, msg...)
	return
}

// Default logger utilizes os.Args[0] for name, os.Stdout as writer, with warn severity.
var Default = Logger{
	": " + filepath.Base(os.Args[0]) + ":",
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
	pm := Default.Name() + Sname[Sfatal]
	if err != nil {
		pm += err.Error()
	}
	panic(pm)
}
