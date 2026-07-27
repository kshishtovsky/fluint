package term

// EventType discriminates the InputEvent tagged union.
type EventType uint8

const (
	// EventNone is the zero value — an invalid/empty event.
	EventNone EventType = iota
	// EventKey represents a keyboard input event.
	EventKey
	// EventMouse represents a mouse input event.
	EventMouse
)

// KeyCode identifies special keys that have no rune representation.
type KeyCode uint16

const (
	KeyNone       KeyCode = iota // Regular character — use KeyEvent.Rune.
	KeyEscape                    // ESC
	KeyEnter                     // Enter / Return
	KeyTab                       // Tab
	KeyBacktab                   // Shift+Tab (CSI Z)
	KeyBackspace                 // Backspace
	KeyDelete                    // Delete
	KeyInsert                    // Insert
	KeyUp                        // Arrow up
	KeyDown                      // Arrow down
	KeyLeft                      // Arrow left
	KeyRight                     // Arrow right
	KeyHome                      // Home
	KeyEnd                       // End
	KeyPageUp                    // Page Up
	KeyPageDown                  // Page Down
	KeyF1                        // F1
	KeyF2                        // F2
	KeyF3                        // F3
	KeyF4                        // F4
	KeyF5                        // F5
	KeyF6                        // F6
	KeyF7                        // F7
	KeyF8                        // F8
	KeyF9                        // F9
	KeyF10                       // F10
	KeyF11                       // F11
	KeyF12                       // F12
	KeyPasteStart                // Bracketed paste begin marker (CSI 200~)
	KeyPasteEnd                  // Bracketed paste end marker (CSI 201~)
)

// KeyMod represents keyboard modifier flags. Multiple modifiers are
// combined with bitwise OR.
type KeyMod uint8

const (
	ModShift KeyMod = 1 << iota // Shift
	ModAlt                      // Alt / Option
	ModCtrl                     // Control
	ModMeta                     // Meta / Super / Windows
)

// KeyEvent represents a keyboard input event. For printable characters,
// Code is KeyNone and Rune holds the character. For special keys, Code
// identifies the key and Rune is 0.
type KeyEvent struct {
	Rune rune    // Character, or 0 for special keys.
	Code KeyCode // Special key code, or KeyNone for printable chars.
	Mod  KeyMod  // Active modifier keys.
}

// MouseButton identifies a mouse button.
type MouseButton uint8

const (
	MouseNone      MouseButton = iota // No button (used in release events).
	MouseLeft                         // Left button.
	MouseMiddle                       // Middle button.
	MouseRight                        // Right button.
	MouseWheelUp                      // Scroll wheel up.
	MouseWheelDown                    // Scroll wheel down.
)

// MouseAction identifies the type of mouse event.
type MouseAction uint8

const (
	MousePress   MouseAction = iota + 1 // Button press.
	MouseRelease                        // Button release.
	MouseMotion                         // Motion with button held.
)

// MouseEvent represents a mouse input event with position and modifiers.
type MouseEvent struct {
	Button MouseButton // Which button was involved.
	Action MouseAction // Press, release, or motion.
	X      uint16      // 0-based column.
	Y      uint16      // 0-based row.
	Mod    KeyMod      // Active modifier keys.
}

// InputEvent is a tagged union of all terminal input events.
// All fields are value types with no pointers — safe and allocation-free
// when passed by value through channels.
//
// Use the Type field to determine which sub-struct is valid:
//   - EventKey:   Key field is valid.
//   - EventMouse: Mouse field is valid.
type InputEvent struct {
	Type  EventType  // Discriminator.
	Key   KeyEvent   // Valid when Type == EventKey.
	Mouse MouseEvent // Valid when Type == EventMouse.
}
