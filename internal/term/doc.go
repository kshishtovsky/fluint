// Package term provides low-level terminal I/O: a zero-allocation
// escape sequence parser, ring buffer for stdin, terminal capabilities
// detection, and a pre-allocated write buffer for single-syscall output.
//
// All hot-path code in this package targets zero allocations per frame
// in steady state (see Performance & Safety Doctrine §3).
//
// This package is internal and not part of the public API.
package term
