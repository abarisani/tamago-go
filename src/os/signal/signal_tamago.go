// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package signal

import (
	"math"
	"os"
	"syscall"
	"time"
)

var loopG uintptr

// Defined by the runtime package.
func getgp() uintptr

func loop() {
	loopG = getgp()

	for {
		time.Sleep(math.MaxInt64)
		process(syscall.SIGINT)
	}
}

func init() {
	watchSignalLoop = loop
}

const numSig = 8

func signum(sig os.Signal) int {
	switch sig := sig.(type) {
	case syscall.Signal:
		i := int(sig)
		if i < 0 || i >= numSig {
			return -1
		}
		return i
	default:
		return -1
	}
}

func enableSignal(sig int)  {}
func disableSignal(sig int) {}
func ignoreSignal(sig int)  {}

func signalIgnored(sig int) bool {
	return false
}

// Interrupt causes an [os.Interrupt] to be sent on the channel
func Interrupt()

// Waiting returns whether package signal is blocked on [Notify].
func Waiting() bool
