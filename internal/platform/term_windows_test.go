//go:build windows

package platform

import (
	"errors"
	"testing"
)

func TestWindowsTerminal_Stubs(t *testing.T) {
	term, err := New()
	if !errors.Is(err, ErrUnsupportedPlatform) {
		t.Errorf("expected ErrUnsupportedPlatform from New(), got %v", err)
	}
	if term != nil {
		t.Errorf("expected nil Terminal on Windows, got %v", term)
	}

	dummy := &Terminal{}
	if err := dummy.EnterRawMode(); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Errorf("expected ErrUnsupportedPlatform from EnterRawMode(), got %v", err)
	}
	if err := dummy.ExitRawMode(); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Errorf("expected ErrUnsupportedPlatform from ExitRawMode(), got %v", err)
	}
	if _, _, err := dummy.GetSize(); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Errorf("expected ErrUnsupportedPlatform from GetSize(), got %v", err)
	}
	if _, err := dummy.Read(nil); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Errorf("expected ErrUnsupportedPlatform from Read(), got %v", err)
	}
	if _, err := dummy.Write(nil); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Errorf("expected ErrUnsupportedPlatform from Write(), got %v", err)
	}
	if err := dummy.Close(); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Errorf("expected ErrUnsupportedPlatform from Close(), got %v", err)
	}
}
