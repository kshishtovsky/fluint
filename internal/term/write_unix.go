//go:build linux || darwin

package term

import "syscall"

// Flush writes all buffered data to fd in a single write(2) syscall
// and resets the buffer.
//
// Principle 3.2 / 3.5: one write(2) per frame — no io.Writer wrappers,
// no bufio.Writer, no intermediate flushes.
//
// If the OS returns a short write (unlikely for ttys but possible),
// it retries until all data is sent. EINTR is handled transparently.
func (w *Writer) Flush(fd uintptr) error {
	if w.n == 0 {
		return nil
	}
	data := w.buf[:w.n]
	for len(data) > 0 {
		n, err := syscall.Write(int(fd), data)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			w.n = 0
			return err
		}
		data = data[n:]
	}
	w.n = 0
	return nil
}
