//go:build !windows

package herdrsock

import (
	"context"
	"net"
)

func dialLocal(ctx context.Context, path string) (net.Conn, error) {
	var dialer net.Dialer
	return dialer.DialContext(ctx, "unix", path)
}
