// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// func Interrupt()
TEXT ·Interrupt(SB),NOSPLIT|NOFRAME,$0
	MOVW	·loopG(SB), R0
	B	runtime·wakeg(SB)
	RET
