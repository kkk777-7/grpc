package game

type Color int

const (
	Empty Color = iota
	Black
	White
	Wall
	None
)

func ColorToStr(c Color) string {
	switch c {
	case Black:
		return "●"
	case White:
		return "○"
	case Empty:
		return " "
	}
	return ""
}

func OpponentColor(c Color) Color {
	switch c {
	case Black:
		return White
	case White:
		return Black
	}
	panic("invalid state")
}
