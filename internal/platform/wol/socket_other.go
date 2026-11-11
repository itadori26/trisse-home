//go:build !unix

package wol

import "syscall"

func allowBroadcast(_, _ string, _ syscall.RawConn) error { return nil }
