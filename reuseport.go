// Copyright (C) 2013 Max Riveiro
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package reuseport provides a function that returns a net.Listener powered by a net.FileListener with a SO_REUSEPORT option set to the socket.
package reuseport

import (
	"errors"
	"net"
	"os"
	"strconv"
	"syscall"
)

const (
	v4                  = 52 // "4"
	v6                  = 54 // "6"
	unsupportedProtoError = "Only tcp4 and tcp6 are supported"
	filePrefix            = "port."
)

// getSockaddr parses protocol and address and returns implementor syscall.Sockaddr: syscall.SockaddrInet4 or syscall.SockaddrInet6.
func getSockaddr(proto, addr string) (sa syscall.Sockaddr, soType int,proto int, err error) {
	var (
		addr4 [4]byte
		addr6 [16]byte
		ip    *net.TCPAddr
		udpIp	*net.UDPAddr
		mode int // 0=not match,1 = tcp,2=udp
	)
	switch proto {
		case "tcp4":
			mode = 1
		case "tcp6":
			mode = 1
		case "tcp":
			mode = 1
		case "udp4":
			mode = 2
		case "udp6":
			mode = 2
		case "udp":
			mode = 2
		default:
			mode = 0
		
	}
	switch mode {
		case 1://TCP
			ip, err = net.ResolveTCPAddr(proto, addr)
			if err != nil {
				return nil, -1, mode, err
			}
		
			switch proto[len(proto)-1] {
			case v4:
				if ip.IP != nil {
					copy(addr4[:], ip.IP[12:16]) // copy last 4 bytes of slice to array
				}
				return &syscall.SockaddrInet4{Port: ip.Port, Addr: addr4}, syscall.AF_INET, mode, nil
			case v6:
				if ip.IP != nil {
					copy(addr6[:], ip.IP) // copy all bytes of slice to array
				}
				return &syscall.SockaddrInet6{Port: ip.Port, Addr: addr6}, syscall.AF_INET6, mode, nil
			}	
		case 2://UDP
			udpIp, err = net.ResolveUDPAddr(proto, addr)
			if err != nil {
				return nil, -1, err
			}
		
			switch proto[len(proto)-1] {
			case v4:
				if udpIp.IP != nil {
					copy(addr4[:], ip.IP[12:16]) // copy last 4 bytes of slice to array
				}
				return &syscall.SockaddrInet4{Port: ip.Port, Addr: addr4}, syscall.AF_INET, mode, nil
			case v6:
				if udpIp.IP != nil {
					copy(addr6[:], ip.IP) // copy all bytes of slice to array
				}
				return &syscall.SockaddrInet6{Port: ip.Port, Addr: addr6}, syscall.AF_INET6, mode, nil
			}
		default:
			return nil, -1, mode, errors.New(unsupportedProtoError)
	}
}

// NewReusablePortListener returns net.FileListener that created from a file discriptor for a socket with SO_REUSEPORT option.
func NewReusablePortListener(proto, addr string) (l net.Listener, err error) {
	var (
		soType, fd int
		file       *os.File
		sockaddr   syscall.Sockaddr
	)

	if sockaddr, soType, mode, err = getSockaddr(proto, addr); err != nil {
		return nil, err
	}
	var IPProto int
	if mode = 1 {
		IPProto = syscall.IPPROTO_TCP
	} else {
		IPProto = syscall.IPPROTO_UDP
	}
	if fd, err = syscall.Socket(soType, syscall.SOCK_STREAM, IPProto); err != nil {
		return nil, err
	}

	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, reusePort, 1); err != nil {
		return nil, err
	}

	if err = syscall.Bind(fd, sockaddr); err != nil {
		return nil, err
	}

	// Set backlog size to the maximum
	if err = syscall.Listen(fd, syscall.SOMAXCONN); err != nil {
		return nil, err
	}

	// File Name get be nil
	file = os.NewFile(uintptr(fd), filePrefix+strconv.Itoa(os.Getpid()))
	if l, err = net.FileListener(file); err != nil {
		return nil, err
	}

	if err = file.Close(); err != nil {
		return nil, err
	}

	return l, err
}
