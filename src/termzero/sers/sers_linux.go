// +build linux

//
// 	Copyright 2015 Martin Capitanio <capnm@capitanio.org>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sers

// extern int setbaudrate(int fd, unsigned int br);
import "C"

func (bp *baseport) SetBaudRate(br uint32) error {
	_, err := C.setbaudrate(C.int(bp.f.Fd()), C.uint(br))
	if err != nil {
		return err
	}
	return nil
}
