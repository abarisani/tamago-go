// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

// textAddrSymName allows to statically initialize the start address of text
// symbols, in a manner equivalent to "-T" flag) when GOOSPKG is used.
const textAddrSymName = "internal/runtime/goospkg.TextAddr"

// setTextAddrFromSym overrides FlagTextAddr with the static initializer of
// textAddrSymName, if the program defines it. Must run after loadlib and
// before textaddress.
func (ctxt *Link) setTextAddrFromSym() {
	ldr := ctxt.loader
	s := ldr.Lookup(textAddrSymName, 0)
	if s == 0 {
		return
	}
	switch data := ldr.Data(s); len(data) {
	case 0:
		return
	case 4:
		*FlagTextAddr = int64(ctxt.Arch.ByteOrder.Uint32(data))
	case 8:
		*FlagTextAddr = int64(ctxt.Arch.ByteOrder.Uint64(data))
	default:
		Exitf("%s: want a 4- or 8-byte integer variable, got %d bytes",
			textAddrSymName, len(data))
	}
}
