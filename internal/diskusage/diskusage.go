package diskusage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const blockSize = 512

// Result holds the outcome of a disk usage walk.
type Result struct {
	UsedBytes   int64
	InodesUsed  int64
	InodesTotal int64
	InodesFree  int64
}

// Walker measures directory usage du-style.
type Walker struct {
	timeout time.Duration
}

// New creates a Walker with the given per-walk timeout.
func New(timeout time.Duration) *Walker {
	return &Walker{timeout: timeout}
}

// Measure walks root and returns used bytes and inode count.
func (w *Walker) Measure(ctx context.Context, root string) (Result, error) {
	ctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	info, err := os.Lstat(root)
	if err != nil {
		return Result{}, fmt.Errorf("stat root %q: %w", root, err)
	}
	if !info.IsDir() {
		return Result{}, fmt.Errorf("root %q is not a directory", root)
	}

	var st syscall.Statfs_t
	if err := syscall.Statfs(root, &st); err != nil {
		return Result{}, fmt.Errorf("statfs %q: %w", root, err)
	}
	inodesTotal := int64(st.Files)
	inodesFree := int64(st.Ffree)

	dev := deviceID(info.Sys())
	seen := make(map[inodeKey]struct{})
	var usedBytes, inodes int64

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		sys := info.Sys()
		if sys == nil {
			return nil
		}

		if deviceID(sys) != dev {
			return filepath.SkipDir
		}

		stat, ok := sys.(*syscall.Stat_t)
		if !ok {
			return nil
		}

		key := inodeKey{dev: deviceID(sys), ino: inodeNumber(stat)}
		if _, exists := seen[key]; exists {
			return nil
		}
		seen[key] = struct{}{}

		usedBytes += stat.Blocks * blockSize
		inodes++

		return nil
	})
	if err != nil {
		return Result{}, fmt.Errorf("walk %q: %w", root, err)
	}

	return Result{
		UsedBytes:   usedBytes,
		InodesUsed:  inodes,
		InodesTotal: inodesTotal,
		InodesFree:  inodesFree,
	}, nil
}

type inodeKey struct {
	dev uint64
	ino uint64
}

func deviceID(sys any) uint64 {
	if stat, ok := sys.(*syscall.Stat_t); ok {
		return uint64(stat.Dev) //nolint:unconvert // Dev type varies by platform.
	}
	return 0
}

func inodeNumber(stat *syscall.Stat_t) uint64 {
	return uint64(stat.Ino)
}
