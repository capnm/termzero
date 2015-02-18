//
// 	Copyright 2015 Martin Capitanio <capnm@capitanio.org>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
//
// http://stackoverflow.com/questions/12646324/how-to-set-a-custom-baud-rate-on-linux
//
#include <stdio.h>
#include <asm/termios.h>

static struct termios2 tio = {0};

int setbaudrate(int fd, unsigned int speed)
{
	int ret = ioctl(fd, TCGETS2, &tio);
	if (ret < 0) return ret;

	tio.c_cflag &= ~CBAUD;
	tio.c_cflag |= BOTHER;
	tio.c_ispeed = speed;
	tio.c_ospeed = speed;
	return ioctl(fd, TCSETS2, &tio);
}

/*

http://nxr.netbsd.org/xref/src/lib/libc/termios/cfmakeraw.c

struct termios2 {
	tcflag_t c_iflag;		// input mode flags
	tcflag_t c_oflag;		// output mode flags
	tcflag_t c_cflag;		// control mode flags
	tcflag_t c_lflag;		// local mode flags
	cc_t c_line;			// line discipline
	cc_t c_cc[NCCS];		// control characters
	speed_t c_ispeed;		// input speed
	speed_t c_ospeed;		// output speed
};

Input flags - software input processing

	IGNBRK	ignore BREAK condition
	BRKINT	map BREAK to SIGINT
	IGNPAR	ignore (discard) parity errors
	PARMRK	mark parity and framing errors
	INPCK	enable checking of parity errors
	ISTRIP	strip 8th bit off chars
	INLCR	map NL into CR
	IGNCR	ignore CR
	ICRNL	map CR to NL (ala CRMOD)
	IXON	enable output flow control
	IXOFF	enable input flow control

Output flags - software output processing

	OPOST   enable following output processing

"Local" flags - dumping ground for other state

	ISIG    enable signals INT, QUIT, [D]SUSP
	ICANON  enable erase, kill, werase, and rprnt special characters
	ECHO    enable echoing
	ECHONL  echo NL even if ECHO is off
	IEXTEN  enable DISCARD and LNEXT special characters

Control flags - hardware control of terminal

	CSIZE   character size mask
	PARENB  parity enable
	CS8     8 bits

man stty: raw same as

-ignbrk -brkint -ignpar -parmrk -inpck -istrip
-inlcr -igncr -icrnl -ixon -ixoff

?? -iuclc -ixany -imaxbel

-opost

-isig -icanon
?? -xcase

min 1 time 0
+need: -echo*

 */

// Disable as much crap as possible.
void setraw(struct termios2* t2)
{
	// input mode flags
	t2->c_iflag &= ~(IGNBRK|BRKINT|IGNPAR|PARMRK|INPCK|ISTRIP|
		INLCR|IGNCR|ICRNL|IXON|IXOFF);
	// output mode flags
	t2->c_oflag &= ~OPOST;
	// local mode flags
	t2->c_lflag &= ~(ISIG|ICANON|ECHO|ECHONL|IEXTEN);
	// control mode flags
	t2->c_cflag &= ~(CSIZE|PARENB);
	t2->c_cflag |= CS8;
	// control characters - wait for 1 byte
	t2->c_cc[VTIME] = 0;
	t2->c_cc[VMIN] = 1;
}
