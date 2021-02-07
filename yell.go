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

const (
	// log severity levels
	Sinfo Severity = iota
	Swarn
	Serror
	Sfatal
	Snolog // disables logging
)

// Sname is the list of severity names (in increasing severity) that appear in logs
var Sname = [...]string{"info", "warn", "error", "fatal"}

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
//  // log to any io.Writer, here warn or higher severity. Update
//  // Logger.Slevel to change log severity level or disable logging.
//  var Logger = yell.Logger{"mypkg", os.Stdout, yell.Swarn}
//
//  // Info tries to log message with info severity
//  func Info(msg string) error {
//  	return Logger.Log(yell.Sinfo, msg)
//  }
//
//  // Warn tries to log message with warn severity
//  func Warn(msg string) error {
//  	return Logger.Log(yell.Swarn, msg)
//  }
//
//  // Error tries to log message with error severity
//  func Error(msg string) (err error) {
//  	err = Logger.Log(yell.Serror, msg)
//  	// extra stuff for error severity
//  	return
//  }
//
//  // Fatal tries to log message with fatal severity
//  func Fatal(msg string) (err error) {
//  	err = Logger.Log(yell.Sfatal, msg)
//  	// probably panic (or exit) in a fatal situation
//  	panic(Logger.Package + ":" + yell.Sname[yell.Sfatal] + ": " + msg)
//  }
type Logger struct {

	// Package or application name. It must not be empty.
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

// Log records message to Logger if level is severe enough for Logger. It tries to
// include log request location in log messages. Use as described in Logger doc.
// If Logger.Writer also implements sync.Locker, Lock/Unlock is used to protect logging.
func (lg *Logger) Log(level Severity, msg string) (err error) {

	if !(lg.Slevel <= level && level < Snolog) {
		return // ignored level
	}
	// prepare all input to Fprintf before locking
	now := time.Now()
	if UTC {
		now = now.UTC()
	}
	ft := now.Format(TimeFormat)

	// try to locate logging location
	_, file, line, fok := runtime.Caller(2)
	if fok {
		file = filepath.Base(file) // full path to file name
	}

	// see if Writer is also a sync.Locker
	if lc, ok := lg.Writer.(locker); ok {

		lc.Lock() // lock just before logging
		defer lc.Unlock()
	}

	if fok {
		_, err = fmt.Fprintf(lg.Writer, "%s: %s:%s: %s:%d: %s\n",
			ft, lg.Package, Sname[level], file, line, msg)
	} else {
		_, err = fmt.Fprintf(lg.Writer, "%s: %s:%s: %s\n",
			ft, lg.Package, Sname[level], msg)
	}
	return
}

// Default logger
var Default = Logger{
	filepath.Base(os.Args[0]), // application name
	os.Stdout, Swarn}          // log to standard output with warn or higher severity

// Info tries to log message with info severity to Default logger
func Info(msg string) error {
	return Default.Log(Sinfo, msg)
}

// Warn tries to log message with warn severity to Default logger
func Warn(msg string) error {
	return Default.Log(Swarn, msg)
}

// Error tries to log message with error severity to Default logger
func Error(msg string) error {
	return Default.Log(Serror, msg)
}

// Fatal tries to log message with fatal severity to Default logger and panics
func Fatal(msg string) (err error) {
	err = Default.Log(Sfatal, msg)
	panic(Default.Package + ":" + Sname[Sfatal] + ": " + msg)
}
