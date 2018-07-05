// +build !mips,!mipsle

package spi

const (
	iocNRBITS   = 8
	iocTYPEBITS = 8
	iocSIZEBITS = 14
	iocDIRBITS  = 2

	iocNRSHIFT   = 0
	iocTYPESHIFT = iocNRSHIFT + iocNRBITS
	iocSIZESHIFT = iocTYPESHIFT + iocTYPEBITS
	iocDIRSHIFT  = iocSIZESHIFT + iocSIZEBITS

	iocNONE  = uintptr(0)
	iocWRITE = uintptr(1)
	iocREAD  = uintptr(2)
)
