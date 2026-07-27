//go:build windows

package platform

// Terminal is a stub implementation for Windows.
type Terminal struct{}

// New returns ErrUnsupportedPlatform on Windows.
func New() (*Terminal, error) {
	return nil, ErrUnsupportedPlatform
}

// EnterRawMode returns ErrUnsupportedPlatform on Windows.
func (t *Terminal) EnterRawMode() error {
	return ErrUnsupportedPlatform
}

// ExitRawMode returns ErrUnsupportedPlatform on Windows.
func (t *Terminal) ExitRawMode() error {
	return ErrUnsupportedPlatform
}

// GetSize returns ErrUnsupportedPlatform on Windows.
func (t *Terminal) GetSize() (width, height int, err error) {
	return 0, 0, ErrUnsupportedPlatform
}

// Read returns ErrUnsupportedPlatform on Windows.
func (t *Terminal) Read(p []byte) (int, error) {
	return 0, ErrUnsupportedPlatform
}

// Write returns ErrUnsupportedPlatform on Windows.
func (t *Terminal) Write(p []byte) (int, error) {
	return 0, ErrUnsupportedPlatform
}

// Close returns ErrUnsupportedPlatform on Windows.
func (t *Terminal) Close() error {
	return ErrUnsupportedPlatform
}
