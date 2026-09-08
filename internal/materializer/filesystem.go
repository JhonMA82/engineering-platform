package materializer

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

// ValidateDestination rejects destinations that could escape the project:
// empty, ".", absolute, non-clean or containing ".." segments. Rejection
// (never sanitization) keeps failures explicit and the project layout
// exactly what the plan declared.
func ValidateDestination(dest string) error {
	if strings.TrimSpace(dest) == "" {
		return domain.Filesystem("destination must not be empty")
	}
	if dest == "." {
		return domain.Filesystem(`destination "." is not allowed`)
	}
	if filepath.IsAbs(dest) {
		return domain.Filesystem(fmt.Sprintf("destination %q is absolute (want a relative project path)", dest))
	}
	if strings.Contains(dest, "\\") {
		return domain.Filesystem(fmt.Sprintf("destination %q must use slash separators", dest))
	}
	slash := filepath.ToSlash(dest)
	if filepath.ToSlash(filepath.Clean(dest)) != slash {
		return domain.Filesystem(fmt.Sprintf("destination %q is not clean", dest))
	}
	for _, seg := range strings.Split(slash, "/") {
		if seg == ".." {
			return domain.Filesystem(fmt.Sprintf("destination %q escapes the project", dest))
		}
		if seg == "" {
			return domain.Filesystem(fmt.Sprintf("destination %q is not clean", dest))
		}
	}
	return nil
}

// CheckCollisions rejects duplicate and nested destinations so two
// components can never claim overlapping subtrees.
func CheckCollisions(dests []string) error {
	for i, a := range dests {
		for j, b := range dests {
			if i == j {
				continue
			}
			if a == b {
				return domain.Materialization(fmt.Sprintf("destination collision: %q claimed twice", a))
			}
			if isNested(a, b) {
				return domain.Materialization(fmt.Sprintf("destination collision: %q is nested inside %q", b, a))
			}
		}
	}
	return nil
}

func isNested(parent, child string) bool {
	return strings.HasPrefix(child, parent+"/")
}

// CopyTree copies src into dst without following symlinks: a symlink whose
// resolved target escapes src aborts the copy (symlink-escape rejection).
// Permission bits (including exec) are preserved; sockets, devices and
// other special files are rejected. Prune paths are removed from dst
// after the copy; missing entries are skipped.
func CopyTree(src, dst string, prune []string) error {
	srcAbs, err := filepath.Abs(src)
	if err != nil {
		return domain.Filesystem(fmt.Sprintf("resolve source: %v", err))
	}
	err = filepath.WalkDir(srcAbs, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return domain.Filesystem(fmt.Sprintf("walk source: %v", walkErr))
		}
		rel, err := filepath.Rel(srcAbs, path)
		if err != nil {
			return domain.Filesystem(fmt.Sprintf("relativize %q: %v", path, err))
		}
		target := filepath.Join(dst, rel)
		if !insideDir(dst, target) {
			return domain.Filesystem(fmt.Sprintf("copy target %q escapes its destination", rel))
		}
		info, err := d.Info()
		if err != nil {
			return domain.Filesystem(fmt.Sprintf("stat %q: %v", rel, err))
		}
		switch {
		case info.IsDir():
			if err := os.MkdirAll(target, 0o755); err != nil {
				return domain.Filesystem(fmt.Sprintf("create dir %q: %v", rel, err))
			}
			if err := os.Chmod(target, info.Mode().Perm()); err != nil {
				return domain.Filesystem(fmt.Sprintf("chmod %q: %v", rel, err))
			}
			return nil
		case info.Mode()&fs.ModeSymlink != 0:
			return copySymlink(srcAbs, path, rel, target)
		case info.Mode().IsRegular():
			return copyFile(path, target, info.Mode().Perm())
		default:
			return domain.Filesystem(fmt.Sprintf("refusing special file %q (mode %s)", rel, info.Mode()))
		}
	})
	if err != nil {
		return err
	}
	for _, p := range prune {
		if err := prunePath(dst, p); err != nil {
			return err
		}
	}
	return nil
}

func copySymlink(srcAbs, path, rel, target string) error {
	link, err := os.Readlink(path)
	if err != nil {
		return domain.Filesystem(fmt.Sprintf("read symlink %q: %v", rel, err))
	}
	resolved := link
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(filepath.Dir(path), link)
	}
	resolved = filepath.Clean(resolved)
	if !insideDir(srcAbs, resolved) {
		return domain.Filesystem(fmt.Sprintf("symlink %q escapes its source (target %q)", rel, link))
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return domain.Filesystem(fmt.Sprintf("create dir for %q: %v", rel, err))
	}
	_ = os.Remove(target)
	if err := os.Symlink(link, target); err != nil {
		return domain.Filesystem(fmt.Sprintf("create symlink %q: %v", rel, err))
	}
	return nil
}

func copyFile(src, dst string, perm fs.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return domain.Filesystem(fmt.Sprintf("open %q: %v", src, err))
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return domain.Filesystem(fmt.Sprintf("create dir for %q: %v", dst, err))
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return domain.Filesystem(fmt.Sprintf("create %q: %v", dst, err))
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return domain.Filesystem(fmt.Sprintf("copy to %q: %v", dst, err))
	}
	if err := out.Close(); err != nil {
		return domain.Filesystem(fmt.Sprintf("close %q: %v", dst, err))
	}
	return nil
}

func prunePath(dst, p string) error {
	if err := checkSafeRelPath(p); err != nil {
		return err
	}
	target := filepath.Join(dst, filepath.FromSlash(p))
	if !insideDir(dst, target) {
		return domain.Filesystem(fmt.Sprintf("prune path %q escapes its destination", p))
	}
	if _, err := os.Lstat(target); os.IsNotExist(err) {
		return nil
	}
	if err := os.RemoveAll(target); err != nil {
		return domain.Filesystem(fmt.Sprintf("prune %q: %v", p, err))
	}
	return nil
}

func checkSafeRelPath(p string) error {
	if strings.TrimSpace(p) == "" || p == "." {
		return domain.Filesystem(fmt.Sprintf("path %q is not a valid relative entry", p))
	}
	if filepath.IsAbs(p) || strings.Contains(p, "\\") {
		return domain.Filesystem(fmt.Sprintf("path %q must be relative", p))
	}
	for _, seg := range strings.Split(filepath.ToSlash(p), "/") {
		if seg == ".." || seg == "" {
			return domain.Filesystem(fmt.Sprintf("path %q escapes or is not clean", p))
		}
	}
	return nil
}

// insideDir reports whether target sits inside root (root itself counts).
func insideDir(root, target string) bool {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	if targetAbs == rootAbs {
		return true
	}
	return strings.HasPrefix(targetAbs, rootAbs+string(os.PathSeparator))
}

// ListFiles returns the sorted slash-separated relative paths of every
// non-directory entry under root for the project manifest.
func ListFiles(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return domain.Filesystem(fmt.Sprintf("list files: %v", walkErr))
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return domain.Filesystem(fmt.Sprintf("relativize %q: %v", path, err))
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ensureEmptyOrNew verifies the output directory is new or an empty dir.
func ensureEmptyOrNew(dir string) error {
	st, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return domain.Filesystem(fmt.Sprintf("stat output dir: %v", err))
	}
	if !st.IsDir() {
		return domain.Filesystem(fmt.Sprintf("output %q exists and is not a directory", dir))
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return domain.Filesystem(fmt.Sprintf("read output dir: %v", err))
	}
	if len(entries) > 0 {
		return domain.Materialization(fmt.Sprintf(
			"refusing to materialize into non-empty directory %q (need an empty or new directory)", dir))
	}
	return nil
}
