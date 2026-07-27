//go:build linux

package platform

import (
	"os"
	"strconv"
	"testing"

	"golang.org/x/sys/unix"
)

func createPty(t *testing.T) (master, slave *os.File) {
	t.Helper()
	mFile, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		t.Skipf("failed to open /dev/ptmx: %v", err)
	}
	mFd := int(mFile.Fd())

	unlock := 0
	if err := unix.IoctlSetInt(mFd, unix.TIOCSPTLCK, unlock); err != nil {
		_ = mFile.Close()
		t.Skipf("unlockpt failed: %v", err)
	}

	ptyNum, err := unix.IoctlGetInt(mFd, unix.TIOCGPTN)
	if err != nil {
		_ = mFile.Close()
		t.Skipf("ptsname failed: %v", err)
	}

	ptsPath := "/dev/pts/" + strconv.Itoa(ptyNum)
	sFile, err := os.OpenFile(ptsPath, os.O_RDWR, 0)
	if err != nil {
		_ = mFile.Close()
		t.Skipf("failed to open slave pty %s: %v", ptsPath, err)
	}

	return mFile, sFile
}
