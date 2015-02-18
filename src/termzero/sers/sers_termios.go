// +build darwin linux
//
// 	Copyright 2012 Michael Meier.
// 	Copyright 2015 Martin Capitanio.
//
// All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sers

/*
#include <stdio.h>
#include <asm/termios.h>

extern void setraw(struct termios2* t2);
extern int ioctl(int i, unsigned int r, void *d);
*/
import "C"

import (
	"os"
	"syscall"
	"unsafe"
)

type baseport struct {
	f *os.File
}

// TODO(capnm) implement this on the C side
func TakeOver(f *os.File) (SerialPort, error) {
	if f == nil {
		return nil, &ParameterError{"f", "needs to be non-nil"}
	}
	bp := &baseport{f}

	tio, err := bp.getattr()
	if err != nil {
		return nil, &Error{"bevore putting fd in raw mode", err}
	}

	// was C.cfmakeraw(tio)
	C.setraw(tio)

	err = bp.setattr(tio)
	if err != nil {
		return nil, &Error{"after putting fd in raw mode", err}
	}

	return bp, nil
}

func (bp *baseport) Read(b []byte) (int, error) {
	return bp.f.Read(b)
}

func (b *baseport) Close() error {
	return b.f.Close()
}

func (bp *baseport) Write(b []byte) (int, error) {
	return bp.f.Write(b)
}

// TODO(capnm) drop
func (bp *baseport) getattr() (*C.struct_termios2, error) {
	var tio C.struct_termios2
	res, err := C.ioctl(C.int(bp.f.Fd()), C.TCGETS2, unsafe.Pointer(&tio))
	if res != 0 || err != nil {
		return nil, err
	}

	return &tio, nil
}

// TODO(capnm) drop
func (bp *baseport) setattr(tio *C.struct_termios2) error {
	res, err := C.ioctl(C.int(bp.f.Fd()), C.TCSETS2, unsafe.Pointer(tio))
	if res != 0 || err != nil {
		return err
	}

	return nil
}

// TODO(capnm) implement this on the C side
func (bp *baseport) SetMode(baudrate, databits, parity, stopbits, handshake uint32) error {
	if baudrate <= 0 {
		return &ParameterError{"baudrate", "has to be > 0"}
	}

	var datamask uint
	switch databits {
	case 5:
		datamask = C.CS5
	case 6:
		datamask = C.CS6
	case 7:
		datamask = C.CS7
	case 8:
		datamask = C.CS8
	default:
		return &ParameterError{"databits", "has to be 5, 6, 7 or 8"}
	}

	if stopbits != 1 && stopbits != 2 {
		return &ParameterError{"stopbits", "has to be 1 or 2"}
	}
	var stopmask uint
	if stopbits == 2 {
		stopmask = C.CSTOPB
	}

	var parmask uint
	switch parity {
	case N:
		parmask = 0
	case E:
		parmask = C.PARENB
	case O:
		parmask = C.PARENB | C.PARODD
	default:
		return &ParameterError{"parity", "has to be N, E or O"}
	}

	var flowmask uint
	switch handshake {
	case NO_HANDSHAKE:
		flowmask = 0
	case RTSCTS_HANDSHAKE:
		flowmask = C.CRTSCTS
	default:
		return &ParameterError{"handshake", "has to be NO_HANDSHAKE or RTSCTS_HANDSHAKE"}
	}

	tio, err := bp.getattr()
	if err != nil {
		return err
	}

	tio.c_cflag &^= C.CSIZE
	tio.c_cflag |= C.tcflag_t(datamask)

	tio.c_cflag &^= C.PARENB | C.PARODD
	tio.c_cflag |= C.tcflag_t(parmask)

	tio.c_cflag &^= C.CSTOPB
	tio.c_cflag |= C.tcflag_t(stopmask)

	tio.c_cflag &^= C.CRTSCTS
	tio.c_cflag |= C.tcflag_t(flowmask)

	if err := bp.setattr(tio); err != nil {
		return err
	}

	if err := bp.SetBaudRate(baudrate); err != nil {
		return err
	}

	return nil
}

// TODO(capnm) implement this on the C side
func (bp *baseport) SetReadParams(minread int, timeout float64) error {
	inttimeout := int(timeout * 10)
	if inttimeout < 0 {
		return &ParameterError{"timeout", "needs to be 0 or higher"}
	}
	// if a timeout is desired but too small for the termios timeout
	// granularity, set the minimum timeout
	if timeout > 0 && inttimeout == 0 {
		inttimeout = 1
	}

	tio, err := bp.getattr()
	if err != nil {
		return err
	}

	tio.c_cc[C.VMIN] = C.cc_t(minread)
	tio.c_cc[C.VTIME] = C.cc_t(inttimeout)

	//fmt.Printf("baud rates from termios: %d, %d\n", tio.c_ispeed, tio.c_ospeed)

	err = bp.setattr(tio)
	if err != nil {
		return err
	}

	return nil
}

var Bauds = map[C.speed_t]uint32{
	syscall.B50:      50,
	syscall.B75:      75,
	syscall.B110:     110,
	syscall.B134:     134,
	syscall.B150:     150,
	syscall.B200:     200,
	syscall.B300:     300,
	syscall.B600:     600,
	syscall.B1200:    1200,
	syscall.B1800:    1800,
	syscall.B2400:    2400,
	syscall.B4800:    4800,
	syscall.B9600:    9600,
	syscall.B19200:   19200,
	syscall.B38400:   38400,
	syscall.B57600:   57600,
	syscall.B115200:  115200,
	syscall.B230400:  230400,
	syscall.B460800:  460800,
	syscall.B500000:  500000,
	syscall.B576000:  576000,
	syscall.B921600:  921600,
	syscall.B1000000: 1000000,
	syscall.B1152000: 1152000,
	syscall.B1500000: 1500000,
	syscall.B2000000: 2000000,
	syscall.B2500000: 2500000,
	syscall.B3000000: 3000000,
	syscall.B3500000: 3500000,
	syscall.B4000000: 4000000,
}

func (bp *baseport) Baudrate() (uint32, uint32) {

	tio, err := bp.getattr()
	if err != nil {
		return 0, 0
	}
	//if bauds[tio.c_ispeed] != 0 {
	//	return Bauds[tio.c_ispeed], Bauds[tio.c_ospeed]
	//}
	return uint32(tio.c_ispeed), uint32(tio.c_ospeed)
}

func Open(fn string) (SerialPort, error) {
	// the order of system calls is taken from Apple's SerialPortSample
	// open the TTY device read/write, nonblocking, i.e. not waiting
	// for the CARRIER signal and without the TTY controlling the process
	f, err := os.OpenFile(fn, syscall.O_RDWR|
		syscall.O_NONBLOCK|
		syscall.O_NOCTTY, 0666)
	if err != nil {
		return nil, err
	}

	s, err := TakeOver(f)
	if err != nil {
		return nil, err
	}

	fd := f.Fd()
	if err = syscall.SetNonblock(int(fd), false); err != nil {
		f.Close()
		return nil, &Error{"putting fd into non-blocking mode", err}
	}

	return s, nil
}
