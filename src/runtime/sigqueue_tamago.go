// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements runtime support for signal handling.

package runtime

import (
	"internal/runtime/math"
	"unsafe"
)

var (
	loopG uintptr
	sig   int
)

//go:linkname signal_loop_init os/signal.signal_loop_init
func signal_loop_init() {
	loopG = uintptr(unsafe.Pointer(getg()))
}

// Called to receive a signal.
// Must only be called from a single goroutine at a time.
//
//go:linkname signal_recv os/signal.signal_recv
func signal_recv() (s int) {
	// Sleep indefinitely until woken up by
	// internal∕runtime∕goospkg.SendSignal
	timeSleep(math.MaxInt64)

	s = sig
	sig = -1

	return
}

// sigsend delivers a signal to the os/signal package, it is implemented in
// sys_tamago_$GOARCH.s, this declearation generates its ABI wrapper.
//
//go:linkname sigsend internal/runtime/goospkg.SendSignal
//go:nosplit
func sigsend(s int)

// signal_waiting returns whether package os/signal is blocked waiting for an
// incoming signal or it is handling one.
//
//go:linkname signal_waiting internal/runtime/goospkg.SignalReady
//go:nosplit
func signal_waiting() bool

// This is used to ensure that we do not drop a signal notification due
// to a race between disabling a signal and receiving a signal.
// This assumes that signal delivery has already been disabled for
// the signal(s) in question, and here we are just waiting to make sure
// that all the signals have been delivered to the user channels
// by the os/signal package.
//
//go:linkname signalWaitUntilIdle os/signal.signalWaitUntilIdle
func signalWaitUntilIdle() {
	for !signal_waiting() {
		Gosched()
	}
}
