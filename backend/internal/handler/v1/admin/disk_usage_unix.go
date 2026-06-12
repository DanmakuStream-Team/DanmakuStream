//go:build !windows

package admin

import (
	"os"
	"path/filepath"
	"syscall"
)

func diskUsage(path string) diskStat {
	var stat syscall.Statfs_t
	target := path
	if _, err := os.Stat(target); err != nil {
		target = filepath.Dir(target)
	}
	if err := syscall.Statfs(target, &stat); err != nil {
		return diskStat{}
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free
	percent := 0.0
	if total > 0 {
		percent = float64(used) / float64(total) * 100
	}
	return diskStat{used: used, total: total, free: free, percent: percent}
}
