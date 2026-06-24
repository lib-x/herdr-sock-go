//go:build windows

package herdrsock

import (
	"context"
	"net"
	"strings"

	"github.com/Microsoft/go-winio"
)

func dialLocal(ctx context.Context, path string) (net.Conn, error) {
	return winio.DialPipeContext(ctx, windowsPipeName(path))
}

func windowsPipeName(path string) string {
	if strings.HasPrefix(path, `\\.\pipe\`) {
		return path
	}
	return `\\.\pipe\` + path
}
