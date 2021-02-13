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

func (m *myWriter) Write(p []byte) (n int, err error) {
	m.n++
	m.wr = m.n // record Write's call order

	if strings.Index(string(p), ":warn: yell_test.go:") < 0 {
		err = errors.New("Write's input missing necessary info")
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

func TestWL(t *testing.T) {
	var wl myLocker
	Default.Writer = &wl.myWriter // only writer

	// Default level = warn, Info() logs?
	if err := Info("msg1", 1.2); err != nil {
		t.Fatal(err)
	}
	if !wl.isZero() {
		t.Fatal("must not log info")
	}

	// log empty list?
	if err := Warn(); err != nil {
		t.Fatal(err)
	}
	if !wl.isZero() {
		t.Fatal("must not log empty list")
	}

	// only calling Write() ?
	if err := Warn("msg2", 3); err != nil {
		t.Fatal(err)
	}
	if wl.wr == 0 || wl.lo != 0 || wl.ul != 0 {
		t.Fatal("writer did not work")
	}

	wl.zero()
	Default.Writer = &wl // writer & locker

	// also calling Lock()/Unlock() ?
	if err := Warn("msg3", true); err != nil {
		t.Fatal(err)
	}
	if !(1 == wl.lo && wl.lo < wl.wr && wl.wr < wl.ul) {
		t.Fatal("writer-locker did not work")
	}
}
