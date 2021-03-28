/*	Copyright (c) 2021, Serhat Şevki Dinçer.
	This Source Code Form is subject to the terms of the Mozilla Public
	License, v. 2.0. If a copy of the MPL was not distributed with this
	file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package yell

import (
	"errors"
	"strings"
	"testing"
)

type myWriter struct {
	n, wr uint32
}

var (
	fileName = "yell_test.go:"
	sevName  = "info"
)

const logName = "yell.test:"

func (m *myWriter) Write(p []byte) (n int, err error) {
	m.n++
	m.wr = m.n // record Write's call order

	if strings.Index(string(p), ": "+logName+sevName+": "+fileName) < 0 {
		err = errors.New("input missing necessary info")
	}
	return
}

type myLocker struct {
	myWriter
	lo, ul uint32
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
	_ = New(":  :", Default.writer, Sinfo)
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
	if Default.Name() != logName {
		t.Fatal("unexpected logger name")
	}

	var wl myLocker
	if !Default.UpdateWriter(&wl.myWriter) { // only writer
		t.Fatal("must update the writer")
	}
	if !wl.isZero() {
		t.Fatal("must not call any method")
	}

	// Default severiy level is warn, Info() logs?
	if err := Info("msg0", 1.2); err != nil {
		t.Fatal(err)
	}
	if !wl.isZero() {
		t.Fatal("must not log info level")
	}

	// set min severity to info
	Default.SetMinLevel(Sinfo)

	// only calling Write() ?
	if err := Info("msg1", 2); err != nil {
		t.Fatal(err)
	}
	if wl.wr != 1 || wl.lo != 0 || wl.ul != 0 {
		t.Fatal("must log info")
	}
	wl.zero()

	// non-positive caller depth must be ignored
	if err := Info(Caller(-1), "msg2", 3); err != nil {
		t.Fatal(err)
	}
	if wl.wr != 1 || wl.lo != 0 || wl.ul != 0 {
		t.Fatal("must log info")
	}
	wl.zero()

	// must not log empty list
	if err := Warn(Caller(1)); err != nil {
		t.Fatal(err)
	}
	if !wl.isZero() {
		t.Fatal("must not log empty list")
	}
	UTC = true // allows UTC time

	if !Default.UpdateWriter(&wl) { // writer & locker
		t.Fatal("must update the writer")
	}
	if !wl.isZero() {
		t.Fatal("must not call any method")
	}

	// also calling Lock()/Unlock() ?
	// Caller(1) will change log location to testing.go:line
	sevName, fileName = "warn", "testing.go:"
	if err := Warn(Caller(1), "msg3", true); err != nil {
		t.Fatal(err)
	}
	if wl.lo != 1 || wl.wr != 2 || wl.ul != 3 {
		t.Fatal("writer-locker did not work")
	}
	wl.zero()

	// test UpdateWriter
	// same writer/locker (for testing)
	if !Default.UpdateWriter(&wl) {
		t.Fatal("must update the writer")
	}
	if wl.wr != 0 || wl.lo != 1 || wl.ul != 2 {
		t.Fatal("must call lock/unlock only")
	}
	wl.zero()

	// different writer/locker
	var wl2 myLocker
	if Default.UpdateWriter(&wl2) {
		t.Fatal("must not update with different locker")
	}
	if !wl.isZero() {
		t.Fatal("must not call lock/unlock")
	}

	// disable logging
	Default.SetMinLevel(Snolog + 1) // to test invalid severity levels
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
}
