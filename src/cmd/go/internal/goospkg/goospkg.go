// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package goospkg implements support for the GOOSPKG build setting (see Go
// proposal #73608).
//
// The GOOSPKG build setting controls which copy of the
// internal/runtime/goospkg source code to use. The default is
// GOROOT/src/internal/runtime/goospkg, but a different implementation can be
// substituted into the build instead.
//
// This package provides the logic needed by the rest of the go command
// to implement the overlay.
//
// When GOOSPKG is empty GOROOT/src/internal/runtime/goospkg is imported as
// expected to resolve [internal/runtime/goospkg].
//
// When GOOSPKG is set it defines a module repository path to be used as alias
// for [internal/runtime/goospkg].
//
// ResolveImport is called to resolve the [internal/runtime/goospkg] import, in
// a manner similar to fips140 snapshot logic (see
// GOROOT/src/cmd/go/internal/fips140).
package goospkg

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"cmd/go/internal/base"
	"cmd/go/internal/cfg"
	"cmd/go/internal/modload"
	"cmd/go/internal/str"
)

const importPath = "internal/runtime/goospkg"

var (
	goosPkgOnce sync.Once
	goosPkgDir  string
	overlayDir  atomic.Pointer[string]
)

// goosPkgSrcDir returns the source directory of the GOOSPKG module.
func goosPkgSrcDir(moduleLoader *modload.Loader) string {
	goosPkgOnce.Do(func() {
		defer func(w bool) { modload.ExplicitWriteGoMod = w }(modload.ExplicitWriteGoMod)
		modload.ExplicitWriteGoMod = true

		for p := cfg.GOOSPKG; p != "" && p != "."; p = path.Dir(p) {
			r, err := modload.ListModules(moduleLoader, context.Background(), []string{p}, 0, "")
			if err != nil || len(r) == 0 || r[0].Error != nil || r[0].Dir == "" {
				continue
			}

			rel := strings.TrimPrefix(cfg.GOOSPKG, r[0].Path)
			goosPkgDir = filepath.Join(r[0].Dir, filepath.FromSlash(rel))

			break
		}

		if len(goosPkgDir) == 0 {
			base.Fatalf("go: GOOSPKG=%q not found in module list", cfg.GOOSPKG)
		}

		overlayDir.Store(&goosPkgDir)
	})

	return goosPkgDir
}

// Overlay reports whether the GOOSPKG overlay replaces the bundled
// internal/runtime/goospkg source.
func Overlay() bool {
	return cfg.Goos == "tamago" && cfg.GOOSPKG != ""
}

// IsOverlayDir reports whether dir holds the GOOSPKG overlay source.
// It reports false until the overlay import has been resolved, which is
// harmless: no package directory can be under an overlay that has not been
// located yet.
func IsOverlayDir(dir string) bool {
	d := overlayDir.Load()
	if d == nil || *d == "" {
		return false
	}
	return str.HasFilePathPrefix(filepath.Clean(dir), *d)
}

// ResolveImport resolves the import path imp.
func ResolveImport(moduleLoader *modload.Loader, imp string) (newPath, dir string, ok bool) {
	if cfg.Goos != "tamago" {
		return "", "", false
	}

	// The GOOSPKG module's own path for the overlay directory is an alias
	// for internal/runtime/goospkg: both resolve to the same directory and
	// are built as a single package, so the runtime and the module see the
	// same variables.
	alias := cfg.GOOSPKG != "" && imp == cfg.GOOSPKG

	if imp != importPath && !alias {
		return "", "", false
	}

	if cfg.GOOSPKG != "" {
		dir = goosPkgSrcDir(moduleLoader)
	} else {
		// fallback to bundled Linux userspace GOOSPKG
		if os.Getenv("GOHOSTOS") == "linux" && (cfg.Goarch == "amd64" || cfg.Goarch == "arm" || cfg.Goarch == "arm64" || cfg.Goarch == "loong64" || cfg.Goarch == "riscv64") {
			dir = filepath.Join(cfg.GOROOT, "src", importPath)
		} else {
			base.Fatalf("go: GOOS %s unsupported without external GOOSPKG on %s/%s", cfg.Goos, os.Getenv("GOHOSTOS"), cfg.Goarch)
		}
	}

	return importPath, dir, true
}

// Visible reports whether path is the GOOSPKG overlay, which is exempt from
// the internal package rule.
func Visible(path string) bool {
	return cfg.Goos == "tamago" && path == importPath
}
