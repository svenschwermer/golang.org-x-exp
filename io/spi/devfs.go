// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux

package spi

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/exp/io/spi/driver"
)

const iocMagic = uintptr('k')

type payload struct {
	tx       uint64
	rx       uint64
	length   uint32
	speed    uint32
	delay    uint16
	bits     uint8
	csChange uint8
	txNBits  uint8
	rxNBits  uint8
	pad      uint16
}

var payloadSize = int(unsafe.Sizeof(payload{}))

// Devfs is an SPI driver that works against the devfs.
// You need to have loaded the "spidev" Linux module to use this driver.
type Devfs struct {
	// Dev is the device to be opened.
	// Device name is usually in the /dev/spidev<bus>.<chip> format.
	// Required.
	Dev string

	// Mode is the SPI mode. SPI mode is a combination of polarity and phases.
	// CPOL is the high order bit, CPHA is the low order. Pre-computed mode
	// values are Mode0, Mode1, Mode2 and Mode3. The value of the mode argument
	// can be overriden by the device's driver.
	// Required.
	Mode Mode

	// MaxSpeed is the max clock speed (Hz) and can be overriden by the device's driver.
	// Required.
	MaxSpeed int64
}

// Open opens the provided device with the specified options
// and returns a connection.
func (d *Devfs) Open() (driver.Conn, error) {
	f, err := os.OpenFile(d.Dev, os.O_RDWR, os.ModeDevice)
	if err != nil {
		return nil, err
	}
	conn := &devfsConn{f: f}
	if err := conn.Configure(driver.Mode, int(d.Mode)); err != nil {
		conn.Close()
		return nil, err
	}
	if err := conn.Configure(driver.MaxSpeed, int(d.MaxSpeed)); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

type devfsConn struct {
	f        *os.File
	mode     uint8
	speed    uint32
	bits     uint8
	delay    uint16
	csChange uint8
}

func (c *devfsConn) Configure(k, v int) error {
	switch k {
	case driver.Mode:
		m := uint8(v)
		if err := c.ioctl(iow(1, 1), uintptr(unsafe.Pointer(&m))); err != nil {
			return fmt.Errorf("error setting mode to %v: %v", m, err)
		}
		c.mode = m
	case driver.Bits:
		b := uint8(v)
		if err := c.ioctl(iow(3, 1), uintptr(unsafe.Pointer(&b))); err != nil {
			return fmt.Errorf("error setting bits per word to %v: %v", b, err)
		}
		c.bits = b
	case driver.MaxSpeed:
		s := uint32(v)
		if err := c.ioctl(iow(4, 4), uintptr(unsafe.Pointer(&s))); err != nil {
			return fmt.Errorf("error setting speed to %v: %v", s, err)
		}
		c.speed = s
	case driver.Order:
		o := uint8(v)
		if err := c.ioctl(iow(2, 1), uintptr(unsafe.Pointer(&o))); err != nil {
			return fmt.Errorf("error setting bit order to %v: %v", o, err)
		}
	case driver.Delay:
		c.delay = uint16(v)
	case driver.CSChange:
		c.csChange = uint8(v)
	default:
		return fmt.Errorf("unknown key: %v", k)
	}
	return nil
}

func (c *devfsConn) Tx(w, r []byte) error {
	if r == nil {
		r = make([]byte, len(w))
	}
	// TODO(jbd): len(w) == len(r)?
	// TODO(jbd): Allow nil w.
	p := payload{
		tx:       uint64(uintptr(unsafe.Pointer(&w[0]))),
		rx:       uint64(uintptr(unsafe.Pointer(&r[0]))),
		length:   uint32(len(w)),
		speed:    c.speed,
		delay:    c.delay,
		bits:     c.bits,
		csChange: c.csChange,
	}
	// TODO(jbd): Read from the device and fill rx.
	return c.ioctl(iocMessage(1), uintptr(unsafe.Pointer(&p)))
}

func (c *devfsConn) Close() error {
	return c.f.Close()
}

// ioc returns the device specific SPI request code for the specified direction,
// number and size to be used in the ioctl call.
// Corresponds to _IOC(dir, SPI_IOC_MAGIC, nr, size).
func ioc(dir uintptr, nr, size int) uintptr {
	return (dir << iocDIRSHIFT) | (iocMagic << iocTYPESHIFT) |
		(uintptr(nr) << iocNRSHIFT) | (uintptr(size) << iocSIZESHIFT)
}

// iow returns the device specific SPI write request code for the specified
// number and size to be used in the ioctl call.
// Corresponds to _IOW(SPI_IOC_MAGIC, nr, size).
func iow(nr, size int) uintptr {
	return ioc(iocWRITE, nr, size)
}

// iocMessage returns the device specific value for the SPI
// message payload to be used in the ioctl call.
// n represents the number of messages.
// Corresponds to SPI_IOC_MESSAGE(n).
func iocMessage(n int) uintptr {
	return iow(0, n*payloadSize)
}

// ioctl makes an IOCTL on the open device file descriptor.
func (c *devfsConn) ioctl(a1, a2 uintptr) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL, c.f.Fd(), a1, a2,
	)
	if errno != 0 {
		return syscall.Errno(errno)
	}
	return nil
}
