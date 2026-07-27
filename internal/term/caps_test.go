package term

import "testing"

func TestDetectColorTerm(t *testing.T) {
	tests := []struct {
		name       string
		envTerm    string
		envColor   string
		envProgram string
		wantDepth  ColorDepth
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
		name            string
		program         string
		wantTrue        bool
		wantKitty       bool
		wantSyncOutput  bool
	}{
		{"kitty", "kitty", true, true, true},
		{"wezterm", "WezTerm", true, false, true},
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
			if caps.HasKittyKbd != tt.wantKitty {
				t.Fatalf("HasKittyKbd = %v, want %v", caps.HasKittyKbd, tt.wantKitty)
			}
			if caps.HasSyncOutput != tt.wantSyncOutput {
				t.Fatalf("HasSyncOutput = %v, want %v", caps.HasSyncOutput, tt.wantSyncOutput)
			}
		})
	}
}

func TestDetectTermProgram(t *testing.T) {
	t.Setenv("TERM", "xterm")
	t.Setenv("COLORTERM", "")
	t.Setenv("TERM_PROGRAM", "MyCustomTerm")

	caps := Detect()
	if caps.TermProgram != "MyCustomTerm" {
		t.Fatalf("TermProgram = %q, want %q", caps.TermProgram, "MyCustomTerm")
	}
}
