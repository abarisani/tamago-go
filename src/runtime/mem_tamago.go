// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	goospkg "internal/runtime/goospkg"
	"unsafe"
)

func sbrk(n uintptr) unsafe.Pointer {
	// Plan 9 sbrk from /sys/src/libc/9sys/sbrk.c
	bl := bloc
	n = memRound(n)
	if bl+n > blocMax {
		switch {
		case goospkg.BlocMax > 0:
			if bl+n > goospkg.BlocMax {
				return nil
			}
		case bl+n > uintptr(g0.stack.lo):
			return nil
		default:
			memclrNoHeapPointers(unsafe.Pointer(bl), n)
		}
		blocMax = bl + n
	}
	bloc += n
	return unsafe.Pointer(bl)
}
