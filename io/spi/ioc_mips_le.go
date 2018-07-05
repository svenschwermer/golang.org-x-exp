// +build mips mipsle

package spi

const (
	iocNRBITS   = 8
	iocTYPEBITS = 8
	iocSIZEBITS = 13
	iocDIRBITS  = 3

	iocNRSHIFT   = 0
	iocTYPESHIFT = iocNRSHIFT + iocNRBITS
	iocSIZESHIFT = iocTYPESHIFT + iocTYPEBITS
	iocDIRSHIFT  = iocSIZESHIFT + iocSIZEBITS

	iocNONE  = uintptr(1)
	iocREAD  = uintptr(2)
	iocWRITE = uintptr(4)
)
