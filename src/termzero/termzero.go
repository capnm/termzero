//
//	Copyright (c) 2015 Martin Capitanio <capnm@capitanio.org>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
//

//
// A simple raw (1:1) mode UART terminal.
// Usefull for example to talk via a RS232 dongle or the RPi SoC UART
// to another on-chip (TI MSP, Atmel Mega/Tiny, ...) USARTs.
//
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	//"os/exec"
	"io"

	"termzero/sers"
)

const (
	//defBaudrate uint = 300
	defBaudrate uint = 38400
	//defBaudrate uint = 250000

	databits  uint32 = 8
	parity    uint32 = sers.N
	stopbits  uint32 = 1
	handshake uint32 = sers.NO_HANDSHAKE
)

func main() {

	var baudrate_flag *uint = flag.Uint("b", defBaudrate, "Baud rate")
	flag.Parse()
	baudrate := uint32(*baudrate_flag)

	fmt.Print("termzero v1.1 - ")

	r := bufio.NewReader(os.Stdin)
	w := bufio.NewWriter(os.Stdout)

	pd := findSerialPortDevice()
	fmt.Print(pd, " - ")
	port, err := sers.Open(pd)
	//port, err := os.Open(pd)
	//port, err := sers.SioOpen(pd)
	if err != nil {
		fmt.Println("Fatal: serial port:", err)
		os.Exit(1)
	}
	defer port.Close()

	bi, bo := port.Baudrate()
	fmt.Printf("baudrate (i/o): %d %d\n", bo, bi)

	err = port.SetMode(baudrate, databits, parity, stopbits, handshake)
	if err != nil {
		fmt.Println("Fatal: setup serial port:", err)
	}

	// done by setting raw
	if false {
		//                     min / time-out
		err = port.SetReadParams(1, 0)
		if err != nil {
			fmt.Println("Fatal: setup serial port:", err)
		}
	}

	bi2, bo2 := port.Baudrate()
	if bi != bi2 && bo != bo2 {
		fmt.Printf("set baudrate to (i/o): %d %d\n", bo2, bi2)
	}

	go readFromPort(w, port)

	for {
		b, err := r.ReadBytes('\n')
		if err != nil {
			fmt.Println("Fatal: stdio read:", err)
			// ctrl+d -> EOF
			os.Exit(0)
		}
		_, err = port.Write(b)
		if err != nil {
			fmt.Println("Fatal: port write:", err)
			os.Exit(1)
		}
	}

}

func readFromPort(w *bufio.Writer, rp io.Reader) {
	b := make([]byte, 256)
	n, err := rp.Read(b)
	if err != nil {
		fmt.Println("Fatal: port read:", err)
		os.Exit(1)
	}
	b2 := b[:n]
	for {
		//w.Write([]byte("."))
		_, err = w.Write(b2)
		if err != nil {
			fmt.Println("Fatal: stdio write:", err)
		}
		w.Flush()
		n, err = rp.Read(b)
		if err != nil {
			fmt.Println("Fatal: port read:", err)
			os.Exit(1)
		}
		b2 = b[:n]
	}
}

var serialPortDevices = []string{
	"/dev/ttyAMA0", // RPi UART
	"/dev/ttyUSB0", // USB UART dongle
	"/dev/ttyUSB1",
	"/dev/ttyUSB2",
	"/dev/ttyUSB3",
}

func findSerialPortDevice() string {
	for _, d := range serialPortDevices {
		if fileExist(d) {
			return d
		}
	}
	return "/dev/ttyUSB0"
}

/*
func setTerm() {
	// disable input buffering
	exec.Command("stty", "-F", termDev, "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/ttyUSB0", "-echo").Run()

	// ...

	// restore the echoing state when exiting
	//defer exec.Command("stty", "-F", "/dev/ttyUSB0", "echo").Run()
}
*/
