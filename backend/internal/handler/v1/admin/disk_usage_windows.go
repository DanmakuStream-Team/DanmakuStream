//go:build windows

package admin

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func diskUsage(path string) diskStat {
	target := path
	if _, err := os.Stat(target); err != nil {
		target = filepath.Dir(target)
	}
	if target == "" {
		target = "."
	}

	abs, err := filepath.Abs(target)
	if err != nil {
		abs = target
	}

	var freeBytesAvailable uint64
	var totalBytes uint64
	var totalFreeBytes uint64
	if err := windows.GetDiskFreeSpaceEx(windows.StringToUTF16Ptr(abs), &freeBytesAvailable, &totalBytes, &totalFreeBytes); err != nil {
		return diskStat{}
	}

	used := totalBytes - totalFreeBytes
	percent := 0.0
	if totalBytes > 0 {
		percent = float64(used) / float64(totalBytes) * 100
	}
	return diskStat{used: used, total: totalBytes, free: totalFreeBytes, percent: percent}
}
