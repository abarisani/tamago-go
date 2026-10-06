// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build 386 || amd64 || arm64 || loong64 || ppc64 || ppc64le || riscv64 || s390x || wasm

package math

var haveArchFloor = (soft != "1")

func archFloor(x float64) float64

var haveArchCeil = (soft != "1")

func archCeil(x float64) float64

var haveArchTrunc = (soft != "1")

func archTrunc(x float64) float64
