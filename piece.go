package main

import "strings"

type Piece int8

const (
	Black Piece = 1 << iota
	White
	InvalidPiece

	Both = Black | White
)

func ParsePiece(s string) Piece {
	s = strings.ReplaceAll(strings.ToLower(s), " ", "_")
	switch s {
	case "none":
		return 0
	case "black":
		return Black
	case "white":
		return White
	case "both", "black_and_white", "white_and_black":
		return Both
	default:
		return InvalidPiece
	}
}

func (p Piece) IsValid() bool {
	return p&^Both > 0
}

func (p Piece) String() string {
	switch p {
	case 0:
		return "None"
	case Black:
		return "Black"
	case White:
		return "White"
	case Both:
		return "Both"
	default:
		return "Invalid"
	}
}

func (p Piece) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *Piece) UnmarshalText(text []byte) error {
	*p = ParsePiece(string(text))
	return nil
}
