//go:build !linux

package gateway

import (
	"errors"

	"trisse-home/internal/modules/system/domain"
)

func statfs(string) (domain.Usage, error) {
	return domain.Usage{}, errors.New("disk statistics are only supported on Linux")
}
