package term

import (
	"bytes"
	"testing"
	"time"
)

type blockingMockTerm struct {
	writeBuf bytes.Buffer
	blockCh  chan struct{}
}

func (m *blockingMockTerm) Read(p []byte) (int, error) {
	<-m.blockCh // block indefinitely until test cleanup
	return 0, nil
}

func (m *blockingMockTerm) Write(p []byte) (int, error) {
	return m.writeBuf.Write(p)
}

type responsiveMockTerm struct {
	writeBuf bytes.Buffer
	response []byte
}

func (m *responsiveMockTerm) Read(p []byte) (int, error) {
	n := copy(p, m.response)
	return n, nil
}

func (m *responsiveMockTerm) Write(p []byte) (int, error) {
	return m.writeBuf.Write(p)
}

func TestDetectColorTerm(t *testing.T) {
	tests := []struct {
		name       string
		envTerm    string
		envColor   string
		envProgram string
		wantDepth  int
	}{
		{
			name:      "truecolor via COLORTERM",
			envTerm:   "xterm-256color",
			envColor:  "truecolor",
			wantDepth: ColorTrue,
		},
		{
			name:      "24bit via COLORTERM",
			envTerm:   "screen",
			envColor:  "24bit",
			wantDepth: ColorTrue,
		},
		{
			name:      "256 via TERM suffix",
			envTerm:   "xterm-256color",
			envColor:  "",
			wantDepth: Color256,
		},
		{
			name:      "dumb terminal",
			envTerm:   "dumb",
			envColor:  "",
			wantDepth: ColorMono,
		},
		{
			name:      "empty TERM",
			envTerm:   "",
			envColor:  "",
			wantDepth: ColorMono,
		},
		{
			name:      "plain xterm defaults to 16",
			envTerm:   "xterm",
			envColor:  "",
			wantDepth: Color16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TERM", tt.envTerm)
			t.Setenv("COLORTERM", tt.envColor)
			t.Setenv("TERM_PROGRAM", tt.envProgram)

			caps := Detect()
			if caps.ColorDepth != tt.wantDepth {
				t.Fatalf("ColorDepth = %d, want %d", caps.ColorDepth, tt.wantDepth)
			}
		})
	}
}

func TestDetectKnownPrograms(t *testing.T) {
	tests := []struct {
		name           string
		program        string
		wantTrue       bool
		wantKitty      bool
		wantSyncOutput bool
	}{
		{"kitty", "kitty", true, true, true},
		{"wezterm", "WezTerm", true, true, true},
		{"ghostty", "ghostty", true, false, true},
		{"iterm", "iTerm.app", true, false, false},
		{"unknown", "unknown-term", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TERM", "xterm")
			t.Setenv("COLORTERM", "")
			t.Setenv("TERM_PROGRAM", tt.program)

			caps := Detect()
			if tt.wantTrue && caps.ColorDepth != ColorTrue {
				t.Fatalf("ColorDepth = %d, want ColorTrue", caps.ColorDepth)
			}
			if caps.KittyKeyboard != tt.wantKitty {
				t.Fatalf("KittyKeyboard = %v, want %v", caps.KittyKeyboard, tt.wantKitty)
			}
			if caps.SyncOutput != tt.wantSyncOutput {
				t.Fatalf("SyncOutput = %v, want %v", caps.SyncOutput, tt.wantSyncOutput)
			}
		})
	}
}

func TestDetectActive_Timeout(t *testing.T) {
	t.Setenv("TERM", "xterm")
	t.Setenv("COLORTERM", "")
	t.Setenv("TERM_PROGRAM", "")

	mock := &blockingMockTerm{
		blockCh: make(chan struct{}),
	}

	start := time.Now()
	caps := DetectActiveTimeout(mock, 100*time.Millisecond)
	elapsed := time.Since(start)

	if elapsed < 90*time.Millisecond || elapsed > 300*time.Millisecond {
		t.Fatalf("DetectActive took %v, expected ~100ms", elapsed)
	}

	if caps.SyncOutput != false {
		t.Fatalf("SyncOutput = %v, want false baseline fallback", caps.SyncOutput)
	}

	if mock.writeBuf.String() != DECRQMSyncOutput {
		t.Fatalf("Sent escape sequence = %q, want %q", mock.writeBuf.String(), DECRQMSyncOutput)
	}
}

func TestDetectActive_SyncOutput(t *testing.T) {
	t.Setenv("TERM", "xterm")
	t.Setenv("COLORTERM", "")
	t.Setenv("TERM_PROGRAM", "")

	tests := []struct {
		name       string
		resp       string
		wantResult bool
	}{
		{"supported mode set (1)", "\x1b[?2026;1$y", true},
		{"supported mode reset (2)", "\x1b[?2026;2$y", true},
		{"unsupported mode (0)", "\x1b[?2026;0$y", false},
		{"invalid response", "\x1b[?9999;1$y", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &responsiveMockTerm{
				response: []byte(tt.resp),
			}
			caps := DetectActiveTimeout(mock, 50*time.Millisecond)
			if caps.SyncOutput != tt.wantResult {
				t.Fatalf("SyncOutput = %v, want %v", caps.SyncOutput, tt.wantResult)
			}
		})
	}
}

func BenchmarkDetectEnv(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = DetectEnv("truecolor", "xterm-256color", "kitty")
	}
}

func BenchmarkDetect(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Detect()
	}
}
