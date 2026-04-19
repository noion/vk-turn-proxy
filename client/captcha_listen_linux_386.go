//go:build linux && 386

package main

// On linux/386 (iSH on iOS), Go's TCP Accept path calls accept4 (syscall 364)
// which iSH does not implement. We create the listening socket via the
// socketcall(2) multiplexer (syscall 102) and accept connections via
// socketcall SYS_ACCEPT (sub-call 5) instead.

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

// socketcall sub-call numbers (from <linux/net.h>)
const (
	_scSocket     uintptr = 1
	_scBind       uintptr = 2
	_scListen     uintptr = 4
	_scAccept     uintptr = 5
	_scSetSockOpt uintptr = 14
)

func doSocketcall(call uintptr, args *[6]uintptr) (uintptr, syscall.Errno) {
	r, _, e := syscall.Syscall(102, call, uintptr(unsafe.Pointer(args)), 0)
	return r, e
}

// socketcallListener is a net.Listener backed by a raw fd and socketcall accept.
type socketcallListener struct {
	fd   int
	addr *net.TCPAddr
}

func (l *socketcallListener) Accept() (net.Conn, error) {
	args := [6]uintptr{uintptr(l.fd), 0, 0}
	r, e := doSocketcall(_scAccept, &args)
	if e != 0 {
		return nil, e
	}
	nfd := int(r)
	// Wrap raw fd as net.Conn via os.NewFile + net.FileConn (duplicates the fd).
	f := os.NewFile(uintptr(nfd), "captcha-conn")
	conn, err := net.FileConn(f)
	f.Close()
	return conn, err
}

func (l *socketcallListener) Close() error {
	return syscall.Close(l.fd)
}

func (l *socketcallListener) Addr() net.Addr {
	return l.addr
}

// platformListener creates a TCP listening socket on addrStr via socketcall
// to avoid the accept4 path used by net.Listen on linux/386.
func platformListener(addrStr string) (net.Listener, error) {
	_, portStr, err := net.SplitHostPort(addrStr)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	// socket(AF_INET, SOCK_STREAM, 0)
	socketArgs := [6]uintptr{syscall.AF_INET, syscall.SOCK_STREAM, 0}
	r, e := doSocketcall(_scSocket, &socketArgs)
	if e != 0 {
		return nil, fmt.Errorf("socket: %w", e)
	}
	fd := int(r)

	optval := int32(1)
	ssoArgs := [6]uintptr{uintptr(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR,
		uintptr(unsafe.Pointer(&optval)), 4}
	doSocketcall(_scSetSockOpt, &ssoArgs) //nolint:errcheck

	// struct sockaddr_in { sa_family(2), port_be(2), addr(4), pad(8) }
	var sa [16]byte
	sa[0] = 2 // AF_INET
	sa[1] = 0
	sa[2] = byte(port >> 8)
	sa[3] = byte(port)
	sa[4] = 127; sa[5] = 0; sa[6] = 0; sa[7] = 1 // 127.0.0.1

	bindArgs := [6]uintptr{uintptr(fd), uintptr(unsafe.Pointer(&sa[0])), 16}
	if _, e := doSocketcall(_scBind, &bindArgs); e != 0 {
		_ = syscall.Close(fd)
		return nil, fmt.Errorf("bind 127.0.0.1:%d: %w", port, e)
	}

	listenArgs := [6]uintptr{uintptr(fd), 8}
	if _, e := doSocketcall(_scListen, &listenArgs); e != 0 {
		_ = syscall.Close(fd)
		return nil, fmt.Errorf("listen: %w", e)
	}

	return &socketcallListener{fd: fd, addr: &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: port}}, nil
}
