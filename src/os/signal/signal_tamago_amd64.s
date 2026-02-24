// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// func Interrupt()
TEXT ·Interrupt(SB),NOSPLIT|NOFRAME,$0
	MOVQ	·loopG(SB), AX
	JMP	runtime·wakeg(SB)
	RET

// func Waiting() bool
TEXT ·Waiting(SB),NOSPLIT|NOFRAME,$0-1
	MOVQ	·loopG(SB), AX
	CALL	runtime·findTimer(SB)
	XORQ	$1, BX
	MOVB	BX, ret+0(FP)
	RET
