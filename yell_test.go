/*	Copyright (c) 2021, Serhat Şevki Dinçer.
	This Source Code Form is subject to the terms of the Mozilla Public
	License, v. 2.0. If a copy of the MPL was not distributed with this
	file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package yell

import (
	"errors"
	"os"
	"strings"
	"testing"
)

type myWriter struct {
	n, wr uint32
}

var (
	fileName   = " yell_test.go:"
	sevName    = Sname[Sinfo]
	errMissing = errors.New("input missing necessary info")
)

const logName = ": yell.test:"

func (m *myWriter) Write(p []byte) (int, error) {
	m.n++
	m.wr = m.n // record Write's call order

	if !strings.Contains(string(p), logName+sevName+fileName) {
		return 0, errMissing
	}
	return len(p), nil
}

type myLocker struct {
	myWriter
	lo, ul uint32
}

func (m *myLocker) onlyWrite() bool {
	return m.lo == 0 && m.wr == 1 && m.ul == 0
}

func (m *myLocker) onlyLock() bool {
	return m.lo == 1 && m.wr == 0 && m.ul == 2
}

func (m *myLocker) all() bool {
	return m.lo == 1 && m.wr == 2 && m.ul == 3
}

func (m *myLocker) zero() {
	m.n, m.wr, m.lo, m.ul = 0, 0, 0, 0
}

func (m *myLocker) isZero() bool {
	return m.wr == 0 && m.lo == 0 && m.ul == 0
}

func (m *myLocker) Lock() {
	m.n++
	m.lo = m.n // record Lock's call order
}

func (m *myLocker) Unlock() {
	m.n++
	m.ul = m.n // record Unlock's call order
}

func newPanics() (ok bool) {
	defer func() {
		if recover() != nil {
			ok = true
		}
	}()
	_ = New("badName", os.Stdout, Sinfo)
	return
}

func fatalPanics() (ok bool) {
	defer func() {
		if recover() != nil {
			ok = true
		}
	}()
	_ = Fatal("msg5")
	return
}

func TestWL(t *testing.T) {
	if Default.Name() != logName[2:] {
		t.Fatal("unexpected logger name")
	}

	var wl myLocker
	if !Default.UpdateWriter(&wl.myWriter) { // only writer
		t.Fatal("must update the writer")
	}
	if !wl.isZero() {
		t.Fatal("must not call any method")
	}

	// Default severiy is warn, Info() must not log
	if err := Info("msg0", 1.2); err != nil {
		t.Fatal(err)
	}
	if !wl.isZero() {
		t.Fatal("Info must not log")
	}

	// must not log empty list, and without error
	if err := Warn(Caller(1)); err != nil {
		t.Fatal(err)
	}
	if !wl.isZero() {
		t.Fatal("must not log empty list")
	}

	// must log with error since sevName is still "info:"
	if err := Warn("msg0", 2); err == nil {
		t.Fatal("must log with error")
	}
	if !wl.onlyWrite() {
		t.Fatal("must call only write")
	}
	wl.zero()

	// set min severity to info
	Default.SetLevel(Sinfo)

	// must call only Write
	if err := Info("msg1", 2); err != nil {
		t.Fatal(err)
	}
	if !wl.onlyWrite() {
		t.Fatal("must call only write")
	}
	wl.zero()

	UTC = true // allows UTC time

	// non-positive caller depth must be ignored, without error
	if err := Info(Caller(-1), "msg2", 3); err != nil {
		t.Fatal(err)
	}
	if !wl.onlyWrite() {
		t.Fatal("must call only write")
	}
	wl.zero()

	// Caller(1) must yield log location as testing.go:line
	fileName = " testing.go:"
	if err := Info(Caller(1), "msg2b", 4); err != nil {
		t.Fatal(err)
	}
	if !wl.onlyWrite() {
		t.Fatal("must call only write")
	}
	wl.zero()

	if !Default.UpdateWriter(&wl) { // writer & locker
		t.Fatal("must update the writer")
	}
	if !wl.isZero() {
		t.Fatal("must not call any method")
	}

	// Warn should log with warn severity, without error now
	sevName = Sname[Swarn]
	if err := Warn(Caller(1), "msg3", true); err != nil {
		t.Fatal(err)
	}
	// must call Lock/Write/Unlock
	if !wl.all() {
		t.Fatal("must call all")
	}
	wl.zero()

	// test UpdateWriter, same writer/locker for testing
	if !Default.UpdateWriter(&wl) {
		t.Fatal("must update the writer")
	}
	if !wl.onlyLock() {
		t.Fatal("must call lock/unlock only")
	}
	wl.zero()

	// different writer/locker
	var wl2 myLocker
	if Default.UpdateWriter(&wl2) {
		t.Fatal("must not update with different locker")
	}
	if !wl.isZero() {
		t.Fatal("must not call any method")
	}

	// disable logging
	Default.SetLevel(Snolog + 1) // to test invalid severity levels
	if Default.GetLevel() != Snolog {
		t.Fatal("must be Snolog")
	}
	if err := Warn("msg4", false); err != nil {
		t.Fatal(err)
	}
	if !wl.isZero() {
		t.Fatal("must not log anything")
	}

	// test Fatal()
	if !fatalPanics() {
		t.Fatal("Fatal() must have panicked")
	}
	if !wl.isZero() {
		t.Fatal("must not log anything")
	}

	// test New()
	if !newPanics() {
		t.Fatal("New() must have panicked")
	}
	if !wl.isZero() {
		t.Fatal("must not log anything")
	}
}
