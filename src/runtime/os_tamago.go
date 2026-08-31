// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build tamago

package runtime

import (
	"internal/abi"
	"internal/runtime/atomic"
	goospkg "internal/runtime/goospkg"
	"internal/runtime/math"
	"unsafe"
)

// CallOnG0 calls a function (func()) on g0 stack.
func CallOnG0(func())

// TextRegion returns the start and end addresses of the physical RAM
// containing the Go runtime executable instructions.
func TextRegion() (start, end uintptr) {
	return firstmoduledata.text, firstmoduledata.etext
}

// DataRegion returns the start and end addresses of the physical RAM
// containing the Go runtime global symbols.
func DataRegion() (start, end uintptr) {
	return firstmoduledata.data, firstmoduledata.enoptrbss
}

type mOS struct {
	waitsemacount uint32
}

func hwinit0() {
	goospkg.InitHW0()
}

func hwinit1() {
	goospkg.InitHW1()
}

func nanotime1() int64 {
	return goospkg.Nanotime()
}

// wakeG modifies a goroutine cached timer for time.Sleep (g.timer) to fire as
// soon as possible.
//
// The function is meant to be invoked within Go assembly and its arguments
// must be passed through registers rather than on the frame pointer, see
// definition in sys_tamago_$GOARCH.s for details.
func wakeG()

//go:linkname getgp os/signal.getgp
func getgp() (gp uintptr) {
	return uintptr(unsafe.Pointer(getg()))
}

// stubs for unused/unimplemented functionality
type sigset struct{}
type gsignalStack struct{}

func goenvs()                        {}
func sigsave(p *sigset)              {}
func msigrestore(sigmask sigset)     {}
func clearSignalHandlers()           {}
func sigblock(exiting bool)          {}
func unminit()                       {}
func mdestroy(mp *m)                 {}
func setProcessCPUProfiler(hz int32) {}
func setThreadCPUProfiler(hz int32)  {}
func initsig(preinit bool)           {}
func osyield()                       {}
func osyield_no_g()                  {}

const stacksize = 8192 * 1024 // 8192KB

// May run with m.p==nil, so write barriers are not allowed.
//
//go:nowritebarrier
func newosproc(mp *m) {
	if goospkg.Task == nil {
		throw("newosproc: not implemented")
	}

	stack := sysAlloc(stacksize, &memstats.stacks_sys, "HW thread stack")
	if stack == nil {
		writeErrStr(failallocatestack)
		exit(1)
	}

	goospkg.Task(unsafe.Pointer(uintptr(stack)+stacksize), unsafe.Pointer(mp), unsafe.Pointer(mp.g0), unsafe.Pointer(abi.FuncPCABI0(mstart)))
}

// Called to initialize a new m (including the bootstrap m).
// Called on the parent thread (main thread in case of bootstrap), can allocate memory.
func mpreinit(mp *m) {
	mp.gsignal = malg(32 * 1024)
	mp.gsignal.m = mp
}

func getCPUCount() int32 {
	return numCPUStartup
}

func osinit() {
	physPageSize = 4096
	numCPUStartup = 1

	if goospkg.Bloc != 0 {
		bloc = goospkg.Bloc
		blocMax = bloc
	} else {
		initBloc()
	}
}

func readRandom(r []byte) int {
	goospkg.InitRNG()
	goospkg.GetRandomData(r)
	return len(r)
}

func signame(sig uint32) string {
	return ""
}

//go:linkname os_sigpipe os.sigpipe
func os_sigpipe() {
	throw("too many writes on closed pipe")
}

//go:nosplit
func crash() {
	*(*int32)(nil) = 0
}

//go:linkname syscall
func syscall(number, a1, a2, a3 uintptr) (r1, r2, err uintptr) {
	switch number {
	// SYS_WRITE
	case 1:
		r1 := write(a1, unsafe.Pointer(a2), int32(a3))
		return uintptr(r1), 0, 0
	default:
		throw("unexpected syscall")
	}

	return
}

//go:nosplit
func write1(fd uintptr, buf unsafe.Pointer, count int32) int32 {
	if fd != 1 && fd != 2 {
		throw("unexpected fd, only stdout/stderr are supported")
	}

	c := uintptr(count)

	for i := uintptr(0); i < c; i++ {
		p := (*byte)(unsafe.Pointer(uintptr(buf) + i))
		goospkg.WriteConsole(*p)
	}

	return int32(c)
}

//go:linkname syscall_now syscall.now
func syscall_now() (sec int64, nsec int32) {
	sec, nsec, _ = time_now()
	return
}

//go:nosplit
func walltime() (sec int64, nsec int32) {
	nano := nanotime()
	sec = nano / 1000000000
	nsec = int32(nano % 1000000000)
	return
}

//go:nosplit
func usleep(us uint32) {
	wake := nanotime() + int64(us)*1000
	for nanotime() < wake {
	}
}

//go:nosplit
func usleep_no_g(usec uint32) {
	usleep(usec)
}

func exit(code int32) {
	if goospkg.Exit != nil {
		goospkg.Exit(code)
	}

	print("exit with code ", code, " halting\n")

	for {
		// hang forever
	}
}

func exitThread(wait *atomic.Uint32) {
	return
}

//go:nosplit
func semacreate(mp *m) {
}

//go:nosplit
func semasleep(ns int64) int {
	var deadline int64
	var v uint32

	gp := getg()
	addr := &gp.m.waitsemacount

	if v = atomic.Load(addr); v > 0 && atomic.Cas(addr, v, v-1) {
		return 0
	}

	if ns >= 0 {
		deadline = nanotime() + ns
	} else {
		deadline = math.MaxInt64
	}

	for {
		if v = atomic.Load(addr); v > 0 {
			if atomic.Cas(addr, v, v-1) {
				return 0
			}
			continue
		}
		if ns >= 0 {
			if deadline-nanotime() <= 0 {
				return -1
			}
		}
		if goospkg.Idle != nil {
			goospkg.Idle(deadline)
		}
	}
}

//go:nosplit
func semawakeup(mp *m) {
	atomic.Xadd(&mp.waitsemacount, 1)

	if goospkg.Wake != nil {
		goospkg.Wake(mp.procid)
	}
}

const preemptMSupported = false

func preemptM(mp *m) {}

func minit() {
	if goospkg.ProcID == nil {
		return
	}

	gp := getg()
	gp.m.procid = goospkg.ProcID()
}

// Stubs so tests can link correctly. These should never be called.
func open(name *byte, mode, perm int32) int32        { panic("not implemented") }
func closefd(fd int32) int32                         { panic("not implemented") }
func read(fd int32, p unsafe.Pointer, n int32) int32 { panic("not implemented") }

//go:nowritebarrierrec
//go:nosplit
func libpreinit() {}

//go:nowritebarrierrec
//go:nosplit
func newosproc0(stacksize uintptr, fn unsafe.Pointer) {
	throw("bad newosproc0")
}
