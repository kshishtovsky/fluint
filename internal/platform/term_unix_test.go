//go:build unix

package platform

import (
	"errors"
	"testing"

	"golang.org/x/sys/unix"
)

func TestTerminal_EnterExitRawMode(t *testing.T) {
	master, slave := createPty(t)
	defer func() { _ = master.Close() }()

	term, err := NewFromFile(slave)
	if err != nil {
		t.Fatalf("NewFromFile failed: %v", err)
	}
	defer func() { _ = term.Close() }()

	beforeTermios, err := unix.IoctlGetTermios(int(slave.Fd()), ioctlGetTermios)
	if err != nil {
		t.Fatalf("failed to get initial termios: %v", err)
	}

	if err := term.EnterRawMode(); err != nil {
		t.Fatalf("EnterRawMode failed: %v", err)
	}

	rawTermios, err := unix.IoctlGetTermios(int(slave.Fd()), ioctlGetTermios)
	if err != nil {
		t.Fatalf("failed to get raw termios: %v", err)
	}

	if (rawTermios.Lflag & (unix.ECHO | unix.ICANON)) != 0 {
		t.Errorf("expected ECHO and ICANON to be cleared, got Lflag=%x", rawTermios.Lflag)
	}

	if err := term.ExitRawMode(); err != nil {
		t.Fatalf("ExitRawMode failed: %v", err)
	}

	afterTermios, err := unix.IoctlGetTermios(int(slave.Fd()), ioctlGetTermios)
	if err != nil {
		t.Fatalf("failed to get termios after exit raw mode: %v", err)
	}

	if *beforeTermios != *afterTermios {
		t.Errorf("ExitRawMode did not restore termios state properly.\nBefore: %+v\nAfter:  %+v", *beforeTermios, *afterTermios)
	}
}

func TestTerminal_ReadWrite(t *testing.T) {
	master, slave := createPty(t)
	defer func() { _ = master.Close() }()

	term, err := NewFromFile(slave)
	if err != nil {
		t.Fatalf("NewFromFile failed: %v", err)
	}
	defer func() { _ = term.Close() }()

	testData := []byte("hello fluint")
	n, err := term.Write(testData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(testData) {
		t.Fatalf("expected %d written bytes, got %d", len(testData), n)
	}

	buf := make([]byte, 64)
	readN, err := master.Read(buf)
	if err != nil {
		t.Fatalf("master.Read failed: %v", err)
	}
	if string(buf[:readN]) != string(testData) {
		t.Fatalf("expected %q, got %q", testData, buf[:readN])
	}
}

func TestTerminal_GetSize(t *testing.T) {
	master, slave := createPty(t)
	defer func() { _ = master.Close() }()

	term, err := NewFromFile(slave)
	if err != nil {
		t.Fatalf("NewFromFile failed: %v", err)
	}
	defer func() { _ = term.Close() }()

	ws := &unix.Winsize{Row: 40, Col: 100}
	if err := unix.IoctlSetWinsize(int(master.Fd()), unix.TIOCSWINSZ, ws); err != nil {
		t.Skipf("IoctlSetWinsize not supported: %v", err)
	}

	term.updateSize()

	w, h, err := term.GetSize()
	if err != nil {
		t.Fatalf("GetSize failed: %v", err)
	}
	if w != 100 || h != 40 {
		t.Errorf("expected size 100x40, got %dx%d", w, h)
	}
}

func TestTerminal_ZeroAllocations(t *testing.T) {
	master, slave := createPty(t)
	defer func() { _ = master.Close() }()

	term, err := NewFromFile(slave)
	if err != nil {
		t.Fatalf("NewFromFile failed: %v", err)
	}
	defer func() { _ = term.Close() }()

	term.width.Store(80)
	term.height.Store(24)

	allocs := testing.AllocsPerRun(100, func() {
		_, _, _ = term.GetSize()
	})
	if allocs > 0 {
		t.Errorf("GetSize allocated memory: %f allocs/op", allocs)
	}
}

func TestTerminal_GracefulDegradation(t *testing.T) {
	master, slave := createPty(t)

	term, err := NewFromFile(slave)
	if err != nil {
		t.Fatalf("NewFromFile failed: %v", err)
	}

	_ = slave.Close()
	_ = master.Close()

	buf := make([]byte, 10)
	_, readErr := term.Read(buf)
	if readErr == nil || !errors.Is(readErr, ErrTTYLost) {
		t.Errorf("expected ErrTTYLost on closed tty read, got %v", readErr)
	}

	_, writeErr := term.Write([]byte("test"))
	if writeErr == nil || !errors.Is(writeErr, ErrWriteFailed) {
		t.Errorf("expected ErrWriteFailed on closed tty write, got %v", writeErr)
	}

	_ = term.Close()
}
