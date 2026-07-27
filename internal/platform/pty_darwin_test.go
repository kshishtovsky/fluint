//go:build darwin

package platform

import (
	"bytes"
	"os"
	"syscall"
	"testing"
	"unsafe"

	"golang.org/x/sys/unix"
)

func createPty(t *testing.T) (master, slave *os.File) {
	t.Helper()
	mFile, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		t.Skipf("failed to open /dev/ptmx: %v", err)
	}
	mFd := uintptr(mFile.Fd())

	buf := make([]byte, 128)
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, mFd, uintptr(unix.TIOCPTYGNAME), uintptr(unsafe.Pointer(&buf[0])))
	if errno != 0 {
		_ = mFile.Close()
		t.Skipf("failed to get pty name on darwin: %v", errno)
	}

	n := bytes.IndexByte(buf, 0)
	if n < 0 {
		n = len(buf)
	}
	ptsPath := string(buf[:n])

	sFile, err := os.OpenFile(ptsPath, os.O_RDWR, 0)
	if err != nil {
		_ = mFile.Close()
		t.Skipf("failed to open slave pty %s: %v", ptsPath, err)
	}

	return mFile, sFile
}
