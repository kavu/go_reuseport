// Copyright (C) 2016 Max Riveiro
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package reuseport

import (
	"errors"
	"net"
	"os"
	"syscall"
)

var unsupportedUDPProtoError = errors.New("Only udp, udp4, udp6 are supported")

func getUDPSockaddr(proto, addr string) (sa syscall.Sockaddr, soType int, err error) {
	var (
		addr4 [4]byte
		addr6 [16]byte
		udp   *net.UDPAddr
	)

	udp, err = net.ResolveUDPAddr(proto, addr)
	if err != nil && udp.IP != nil {
		return nil, -1, err
	}

	udpVersion, err := determineUDPProto(proto, udp)
	if err != nil {
		return nil, -1, err
	}

	switch udpVersion {
	case "udp4":
		copy(addr4[:], udp.IP[12:16]) // copy last 4 bytes of slice to array

		return &syscall.SockaddrInet4{Port: udp.Port, Addr: addr4}, syscall.AF_INET, nil

	case "udp6":
		copy(addr6[:], udp.IP) // copy all bytes of slice to array

		return &syscall.SockaddrInet6{Port: udp.Port, Addr: addr6}, syscall.AF_INET6, nil
	}

	return nil, -1, unsupportedProtoError
}

func determineUDPProto(proto string, ip *net.UDPAddr) (string, error) {
	// If the protocol is set to "udp", we determine the actual protocol
	// version from the size of the IP address. Otherwise, we use the
	// protcol given to us by the caller.

	if ip.IP.To4() != nil {
		return "udp4", nil
	}

	if ip.IP.To16() != nil {
		return "udp6", nil
	}

	return "", unsupportedUDPProtoError
}

// NewReusablePortListener returns net.FileListener that created from a file discriptor for a socket with SO_REUSEPORT option.
func NewReusablePortUDPListener(proto, addr string) (l net.PacketConn, err error) {
	var (
		soType, fd int
		file       *os.File
		sockaddr   syscall.Sockaddr
	)

	if sockaddr, soType, err = getSockaddr(proto, addr); err != nil {
		return nil, err
	}

	syscall.ForkLock.RLock()
	fd, err = syscall.Socket(soType, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err == nil {
		syscall.CloseOnExec(fd)
	}
	syscall.ForkLock.RUnlock()
	if err != nil {
		syscall.Close(fd)
		return nil, err
	}

	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return nil, err
	}

	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, reusePort, 1); err != nil {
		syscall.Close(fd)
		return nil, err
	}

	if err = syscall.Bind(fd, sockaddr); err != nil {
		syscall.Close(fd)
		return nil, err
	}

	file = os.NewFile(uintptr(fd), getSocketFileName(proto, addr))
	if l, err = net.FilePacketConn(file); err != nil {
		syscall.Close(fd)
		return nil, err
	}

	if err = file.Close(); err != nil {
		syscall.Close(fd)
		return nil, err
	}

	return l, err
}
