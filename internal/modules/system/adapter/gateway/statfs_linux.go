//go:build linux

package gateway

import (
	"fmt"
	"syscall"

	"trisse-home/internal/modules/system/domain"
)

func statfs(path string) (domain.Usage, error) {
	var s syscall.Statfs_t
	if err := syscall.Statfs(path, &s); err != nil {
		return domain.Usage{}, fmt.Errorf("statfs %s: %w", path, err)
	}
	bsize := uint64(s.Bsize) //nolint:gosec // block size is always positive
	total := s.Blocks * bsize
	// Used is total minus *free* (not available), matching `df`.
	return domain.Usage{Total: total, Used: total - s.Bfree*bsize}, nil
}
