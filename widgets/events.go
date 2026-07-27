package widgets

// KeyEvent represents a keyboard input event delivered to widgets.
type KeyEvent struct {
	Rune rune   // Character, or 0 for special keys.
	Code KeyCode // Special key code, or KeyNone for printable chars.
	Mod  KeyMod  // Active modifier keys.
}

// KeyCode identifies special keys that have no rune representation.
type KeyCode uint16

const (
	KeyNone     KeyCode = iota // Regular character — use Rune.
	KeyEscape
	KeyEnter
	KeyTab
	KeyBacktab // Shift+Tab
	KeyBackspace
	KeyDelete
	KeyInsert
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
)

// KeyMod represents keyboard modifier flags.
type KeyMod uint8

const (
	ModShift KeyMod = 1 << iota
	ModAlt
	ModCtrl
)

// MouseButton identifies a mouse button.
type MouseButton uint8

const (
	MouseNone      MouseButton = iota
	MouseLeft
	MouseMiddle
	MouseRight
	MouseWheelUp
	MouseWheelDown
)

// MouseAction identifies the type of mouse event.
type MouseAction uint8

const (
	MousePress   MouseAction = iota + 1
	MouseRelease
	MouseMotion
	MouseEnter
	MouseLeave
)

// MouseEvent represents a mouse input event with position.
type MouseEvent struct {
	Button MouseButton
	Action MouseAction
	X      int
	Y      int
	Mod    KeyMod
}
