package znet

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sohaha/zlsgo/zfile"
)

// resolvePath resolves a user supplied path that may be relative.
//
// Relative paths are resolved against the project directory (the directory of
// the running executable), which is the convention of zfile.RealPath used
// across the framework. When the project relative target does not exist but the
// working directory one does, the latter is used so that setups relying on the
// working directory keep working. The second return value reports that
// fallback.
//
// When neither exists the project relative path is returned, so callers report
// a canonical location in their error messages.
func resolvePath(p string) (string, bool) {
	if p == "" || filepath.IsAbs(p) {
		return p, false
	}

	// Paths that walk upwards describe where the caller is, not where the
	// binary lives, so they stay anchored to the working directory.
	if p == ".." || strings.HasPrefix(p, "../") || strings.HasPrefix(p, "..\\") {
		if cwdPath, err := filepath.Abs(p); err == nil {
			return normalizePath(cwdPath), true
		}
		return p, true
	}

	projectPath := zfile.RealPath(p)
	if pathTargetExists(projectPath, p) {
		return projectPath, false
	}

	if cwdPath, err := filepath.Abs(p); err == nil {
		cwdPath = normalizePath(cwdPath)
		if pathTargetExists(cwdPath, p) {
			return cwdPath, true
		}
	}

	return projectPath, false
}

// normalizePath keeps separators in sync with zfile.RealPath.
func normalizePath(p string) string {
	return strings.Replace(p, "\\", "/", -1)
}

// pathTargetExists reports whether the target behind abs exists. Glob patterns
// are probed through their parent directory because the literal pattern never
// exists on disk. original is the path as the caller wrote it.
func pathTargetExists(abs, original string) bool {
	if _, err := os.Stat(abs); err == nil {
		return true
	}

	if strings.ContainsAny(original, "*?[") {
		_, err := os.Stat(filepath.Dir(abs))
		return err == nil
	}

	return false
}
