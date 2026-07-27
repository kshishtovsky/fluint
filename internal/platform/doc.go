// Package platform provides OS-specific terminal operations behind
// build tags. It abstracts termios, ioctl, signal handling, and
// file descriptor polling for Linux, macOS, and Windows (stub).
//
// This package is internal and not part of the public API.
package platform
