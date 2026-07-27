// theme.go defines the fluint chat UI design tokens.
//
// Hallmark-adapted design system for terminal UI:
//   - Named colour tokens (no inline hex)
//   - Typography hierarchy via Attrs (Bold headers, Dim labels, Normal body)
//   - Contrast-safe pairs (every Fg/Bg combination tested for readability)
package main

import "github.com/kshishtovsky/fluint/style"

// ── Palette ─────────────────────────────────────────────────────────

const (
	Paper      = style.Black
	PaperAlt   = style.DarkGray
	PaperHover = style.Color(0x00333333)

	InkPrimary   = style.White
	InkSecondary = style.LightGray
	InkMuted     = style.Color(0x00666666)

	Accent     = style.Green
	AccentWarm = style.Yellow
	AccentHot  = style.Red
	AccentCool = style.Cyan
)

// ── Typography ──────────────────────────────────────────────────────

var (
	TitleStyle  = style.New().Foreground(AccentCool).Background(Paper).Bold()
	HeaderStyle = style.New().Foreground(Accent).Background(Paper).Bold()
	BodyStyle   = style.New().Foreground(InkPrimary).Background(Paper)
	LabelStyle  = style.New().Foreground(InkSecondary).Background(Paper).Dim()
	StatusStyle = style.New().Foreground(InkMuted).Background(Paper).Dim()
	UserStyle   = style.New().Foreground(AccentWarm).Background(Paper)
)

// ── Input states ────────────────────────────────────────────────────

var (
	InputDefault = style.New().Foreground(Paper).Background(InkPrimary)
	InputFocused = style.New().Foreground(Paper).Background(style.White).Bold()
)

// ── Card styles ─────────────────────────────────────────────────────
// Chat message archetypes:
//   - ThinkingCard: dim border, muted bg — "agent is reasoning"
//   - AnswerCard: bright border — "agent response"
//   - InfoCard: subtle border — "structured info block"
//   - UserCard: accent border — "user message"

var (
	ThinkingCardStyle = style.New().
				Background(PaperAlt).
				Shadow(1, 1, style.ShadowColor).
				Padding(1, 0)

	AnswerCardStyle = style.New().
			Background(PaperAlt).
			Shadow(1, 1, style.ShadowColor).
			Padding(1, 0)

	InfoCardStyle = style.New().
			Background(PaperAlt).
			Shadow(1, 1, style.ShadowColor).
			Padding(1, 0)

	UserCardStyle = style.New().
			Background(PaperAlt).
			Shadow(1, 1, style.ShadowColor).
			Padding(1, 0)

	InputCardStyle = style.New().
			Background(Paper).
			SolidBorder(InkMuted).
			Padding(1, 0)
)
