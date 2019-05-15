package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Position ∈ [1, 225]. 0 for invalid.
type Position uint8

func GetPosition(x, y int, isOffset bool) (Position, error) {
	if isOffset {
		x += PositionOffset
		y += PositionOffset
	}
	if x < 0 || x >= BoardSize || y < 0 || y >= BoardSize {
		return InvalidPosition, NewPositionOutOfRangeError(x, y)
	}
	return Position(x + y*BoardSize + 1), nil
}

func ParsePosition(s string) (Position, error) {
	if s == "" || strings.EqualFold(s, "<nil>") ||
		strings.EqualFold(s, "<invalid position>") {
		return InvalidPosition, nil
	}
	var x, y int
	r, w := utf8.DecodeRuneInString(s)
	r = unicode.ToUpper(r)
	if r >= 'A' && r <= 'Z' {
		x = int(r - 'A')
		yU64, err := strconv.ParseUint(s[w:], 10, 8)
		if err != nil {
			return InvalidPosition, NewUnknownPositionError(s)
		}
		y = int(yU64) - 1
	} else {
		return InvalidPosition, NewUnknownPositionError(s)
	}
	return GetPosition(x, y, false)
}

func (p Position) X() int {
	if p == InvalidPosition {
		return -1
	}
	return int(p-1) % BoardSize
}

func (p Position) Y() int {
	if p == InvalidPosition {
		return -1
	}
	return int(p-1) / BoardSize
}

func (p Position) XOffset() int {
	return p.X() - PositionOffset
}

func (p Position) YOffset() int {
	return p.Y() - PositionOffset
}

func (p Position) String() string {
	if p == InvalidPosition {
		return "<invalid position>"
	}
	x, y := p.X(), p.Y()
	if x < 0 || x >= BoardSize || y < 0 || y >= BoardSize {
		return fmt.Sprintf("<out of range position>(%d, %d)", x, y)
	}
	return fmt.Sprintf("%c%d", 'A'+x, y+1)
}

func (p Position) IsOutOfRange() bool {
	return p < MinPosition || p > MaxPosition
}

func (p Position) Move(x, y int) (Position, error) {
	pX, pY := p.X(), p.Y()
	if p.IsOutOfRange() {
		return InvalidPosition, NewPositionOutOfRangeError(pX, pY)
	}
	return GetPosition(pX+x, pY+y, false)
}
