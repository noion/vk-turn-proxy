//go:build !linux || !386

package main

import "net"

// platformListener is a thin wrapper around net.Listen for non-iSH platforms.
func platformListener(addrStr string) (net.Listener, error) {
	return net.Listen("tcp", addrStr)
}
